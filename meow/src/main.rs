use actix_web::{get, post, web, App, HttpServer, Responder, Result};
use argh::FromArgs;
use url::Url;

mod bip39;

#[get("/")]
async fn index(config: web::Data<AppConfig>) -> impl Responder {
    format!(
        "meow - paste bin\nusage: curl --data-binary @<file> {}\n",
        config.base_url
    )
}

#[post("/")]
async fn paste(body: web::Bytes, config: web::Data<AppConfig>) -> Result<impl Responder> {
    let key = crate::bip39::mnemonic(config.key_size);
    let mut path = std::path::PathBuf::from(&config.data_dir);
    path.push(&key);
    tokio::fs::write(path, body).await?;
    Ok(format!(
        "{}\n",
        config.base_url.join(&key).unwrap().to_string()
    ))
}

#[derive(FromArgs, Clone)]
/// paste bin
struct AppConfig {
    /// address to listen on (default: 127.0.0.1:3000)
    #[argh(option, short = 'l', default = "String::from(\"127.0.0.1:3000\")")]
    listen: String,
    /// base url
    #[argh(option, short = 'b')]
    base_url: Url,
    /// key size (default: 3)
    #[argh(option, short = 's', default = "3")]
    key_size: usize,
    /// data dir
    #[argh(option, short = 'd')]
    data_dir: String,
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let args: AppConfig = argh::from_env();
    let listen = args.listen.clone();
    HttpServer::new(move || {
        App::new()
            .app_data(web::Data::new(args.clone()))
            .service(index)
            .service(paste)
            .service(actix_files::Files::new("/", &args.data_dir))
    })
    .bind(listen)?
    .run()
    .await
}
