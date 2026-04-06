import { ensureDate } from './utils';

export interface Recording {
    id: number;
    program_id: number;
    cid: string;
    image_url: string;
    partial: boolean;
    level: string;
    title: string;
    episode_title: string;
    start: string | Date;
    end: string | Date;
}

export interface SourceStream {
    index: number;
    type?: string;
    desc?: string;
}

export function fixRecording(r: Recording) {
    r.start = ensureDate(r.start);
    r.end = ensureDate(r.end);
}

export interface PendingDownload {
    recordingId: number;
    filename: string;
}

export interface QueueUpdatedEvent {
    queue: PendingDownload[];
}

export interface DownloadStartedEvent {
    filename: string;
}

export interface ProgressUpdatedEvent {
    completed: number;
    elapsed: string;
    remaining: string;
}

export interface DownloadErroredEvent {
    filename: string;
    reason: string;
}

export interface StateUpdatedEvent {
    state: string;
    reason: string;
}

export interface StreamSelectionRequestedEvent {
    streams: Array<SourceStream>;
}

export interface ServerEvent {
    correlation?: string;
    queueUpdated?: QueueUpdatedEvent;
    downloadStarted?: DownloadStartedEvent;
    progressUpdated?: ProgressUpdatedEvent;
    downloadErrored?: DownloadErroredEvent;
    stateUpdated?: StateUpdatedEvent;
    selectStreams?: StreamSelectionRequestedEvent;
}

export interface StreamsSelectedEvent {
    streams: Array<SourceStream>;
}

export interface ClientEvent {
    correlation?: string;
    streamsSelected?: StreamsSelectedEvent;
}
