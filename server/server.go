package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/mux"
	e "github.com/rokeller/zt-dl/exec"
	"github.com/rokeller/zt-dl/zattoo"
)

//go:embed client/dist/assets client/dist/index.html
var content embed.FS

type server struct {
	a   *zattoo.Account
	dlq *downloadQueue
	hub *wsHub

	port   uint16
	outdir string
}

func Serve(
	ctx context.Context,
	a *zattoo.Account,
	outdir string,
	port uint16,
	openWebUI bool,
) error {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	s := &server{
		a:   a,
		hub: newHub(),

		port:   port,
		outdir: outdir,
	}
	srv := s.startHttpServer(ctx, wg)
	if openWebUI {
		open(ctx, fmt.Sprintf("http://localhost:%d/", port))
	}

	// Wait for the context to be cancelled or done.
	<-ctx.Done()
	fmt.Println("Shutting down web server ...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); nil != err {
		fmt.Fprintf(os.Stderr, "Failed to shut down web server: %v\n", err)
	}
	wg.Wait()
	fmt.Println("Web server shut down.")
	return nil
}

func (s *server) startHttpServer(ctx context.Context, wg *sync.WaitGroup) *http.Server {
	sub, err := fs.Sub(content, "client/dist")
	if nil != err {
		fmt.Fprintf(os.Stderr, "failed to get sub fs: %v\n", err)
	}

	r := mux.NewRouter()
	api := r.PathPrefix("/api/").Subrouter()
	s.dlq = newDownloadQueue(s)

	// Start both the goroutine that processes the download queue as well as
	// the events websocket hub.
	go s.dlq.Run(ctx)
	go s.hub.run(ctx)

	AddRecordingsApi(s, api)
	AddQueuesApis(s, api)
	r.PathPrefix("/").Handler(http.FileServer(http.FS(sub)))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: r,

		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		defer wg.Done() // let main know we are done cleaning up

		fmt.Printf("Starting web server on port %d ...\n", s.port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Failed to start web server: %v\n", err)
		}
	}()

	return srv
}

func open(ctx context.Context, url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return e.CmdFactory(ctx, cmd, args...).Start()
}
