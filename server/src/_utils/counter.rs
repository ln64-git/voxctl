// src/_utils/counter.rs

// region: --- imports
use crate::AppState;
use actix_web::{web, HttpResponse, Responder};
use std::sync::Arc;
use tokio::sync::{mpsc, Mutex};
use tokio::time::sleep;
use tokio::time::Duration;
// endregion: --- imports

pub async fn start_counter(nexus: web::Data<Arc<Mutex<AppState>>>) -> impl Responder {
    let mut state = nexus.lock().await;
    if state.running.is_none() {
        let (count_send, mut count_recv) = mpsc::channel(1);
        state.running = Some(count_send);
        tokio::spawn(async move {
            let mut count = 0;
            loop {
                tokio::select! {
                    _ = sleep(Duration::from_secs(1)) => {
                        println!("{}", count);
                        count += 1;
                    },
                    msg = count_recv.recv() => {
                        if msg.is_none() {
                            break;
                        }
                    }
                }
            }
        });
        HttpResponse::Ok().body("Starting counter.")
    } else {
        HttpResponse::Ok().body("Counter already running...")
    }
}

pub async fn stop_counter(nexus: web::Data<Arc<Mutex<AppState>>>) -> impl Responder {
    let mut state = nexus.lock().await;

    if let Some(count_send) = state.running.take() {
        let _ = count_send.send(()).await;
        HttpResponse::Ok().body("Program Stopped.")
    } else {
        HttpResponse::Ok().body("Program not running.")
    }
}
