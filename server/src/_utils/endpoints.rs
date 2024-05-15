use crate::AppState;
pub(crate) use crate::_utils::counter::start_counter;
use crate::_utils::counter::stop_counter;
use actix_web::web;
use actix_web::HttpResponse;
pub(crate) use actix_web::Responder;
use std::sync::Arc;
use std::sync::Mutex;

pub async fn test_endpoint(_nexus: web::Data<Arc<Mutex<AppState>>>) -> impl Responder {
    // let _ = start_counter(nexus.clone()).await;
    // let _ = speak_text("Hello World!", state.playback_send.clone()).await;

    // let state = nexus.lock().await;

    // let _ = speak_ollama(
    //     "What does the name Luke represent?".to_owned(),
    //     state.playback_send.clone(),
    // )
    // .await;

    HttpResponse::Ok().body("Test Complete.")
}

pub fn register_endpoints(cfg: &mut web::ServiceConfig) {
    cfg.route("/start", web::get().to(start_counter))
        .route("/stop", web::get().to(stop_counter));
}
