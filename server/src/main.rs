// region: --- imports
pub mod _utils;

use crate::_utils::endpoints::register_endpoints;
use _utils::log::custom_log;

use actix_web::{web, App, HttpServer};
use server::AppState;
use std::fs::create_dir_all;
use std::sync::Arc;
use tokio::sync::Mutex;
// endregion: --- imports

#[tokio::main]
async fn main() -> std::io::Result<()> {
    // Create the logs directory if it doesn't exist
    create_dir_all("server/logs")?;

    let nexus = Arc::new(Mutex::new(AppState { running: None }));

    let server = HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(nexus.clone()))
            .configure(register_endpoints)
    })
    .bind("127.0.0.1:8080")?;

    custom_log("Server starting...");
    let server_handle = server.run();
    custom_log("Server started, ready for requests.");

    server_handle.await?;
    custom_log("Server stopped.");

    Ok(())
}
