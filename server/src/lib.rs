// lib.rs

// region: --- imports
pub mod _utils;
// use _utils::azure;
// use _utils::ollama;
// use _utils::playback;
use std::collections::VecDeque;
use std::sync::atomic::AtomicBool;
use tokio::sync::mpsc;
// endregion: --- imports

#[derive(Debug)]
pub struct AppState {
    pub running: Option<mpsc::Sender<()>>,
}

impl Clone for AppState {
    fn clone(&self) -> Self {
        AppState {
            running: self.running.as_ref().map(|sender| sender.clone()),
        }
    }
}

type SinkId = usize;

#[derive(Debug, Clone)]
pub enum PlaybackCommand {
    Play(Vec<u8>),
    Pause,
    Stop,
    Resume,
}

pub struct PlaybackManager {
    pub next_id: SinkId,
    pub command_queue: VecDeque<PlaybackCommand>,
    pub is_idle: AtomicBool,
    pub current_sink: Option<SinkId>,
}
