package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/mux"
	e "github.com/rokeller/zt-dl/exec"
	"github.com/rokeller/zt-dl/zattoo"
)

//go:embed client/dist/assets client/dist/index.html
var content embed.FS

func Serve(ctx context.Context, a *zattoo.Account, outdir string, port int) error {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	srv := startHttpServer(ctx, a, outdir, port, wg)
	// open(fmt.Sprintf("http://localhost:%d/", port))

	<-ctx.Done()
	fmt.Println("Shutting down web server ...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	fmt.Println("Web server shut down.")
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func startHttpServer(ctx context.Context, a *zattoo.Account, outdir string, port int, wg *sync.WaitGroup) *http.Server {
	sub, err := fs.Sub(content, "client/dist")
	if nil != err {
		fmt.Printf("failed to get sub fs: %v\n", err)
	}

	r := mux.NewRouter()
	api := r.PathPrefix("/api/").Subrouter()
	dlq := NewDownloadQueue(a)
	go dlq.Run(ctx)
	AddRecordingsApi(a, dlq, outdir, api)
	AdQueueApi(api)
	r.PathPrefix("/").Handler(http.FileServer(http.FS(sub)))
	r.Use(NewLogMiddleware().Func())

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,

		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		defer wg.Done() // let main know we are done cleaning up

		fmt.Printf("Starting web server on port %d ...\n", port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("Failed to start web server: %v\n", err)
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
