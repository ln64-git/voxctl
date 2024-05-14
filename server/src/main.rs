// region: --- imports
pub mod _utils;

use _utils::{
    counter::{start_counter, stop_counter},
    // ollama::speak_ollama,
};
use actix_web::{web, App, HttpServer};
use server::AppState;
use std::sync::Arc;
use tokio::sync::Mutex;
// endregion: --- imports

fn register_endpoints(cfg: &mut web::ServiceConfig) {
    cfg.route("/start", web::get().to(start_counter))
        .route("/stop", web::get().to(stop_counter));
}

#[tokio::main]
async fn main() -> std::io::Result<()> {
    std::env::set_var("RUST_LOG", "actix_web=debug");
    env_logger::init();

    let nexus = Arc::new(Mutex::new(AppState { running: None }));

    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(nexus.clone()))
            .configure(register_endpoints)
    })
    .bind("127.0.0.1:8080")?
    .run()
    .await
}
