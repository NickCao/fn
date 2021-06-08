use actix_web::{get, http::header, post, rt, web, App, HttpResponse, HttpServer, Responder};
use anyhow::Result;
use futures::StreamExt;
use rusoto_core::{credential::StaticProvider, request::HttpClient, ByteStream, Region};
use rusoto_s3::{GetObjectRequest, PutObjectRequest, S3Client, S3};
use std::env;
use std::io::{Error, ErrorKind};
use tokio_stream::wrappers::UnboundedReceiverStream;
use url::Url;

mod bip39;

#[get("/")]
async fn index(config: web::Data<AppConfig>) -> impl Responder {
    HttpResponse::Ok().body(format!(
        "meow - paste bin\nusage: curl --data-binary @<file> {}\n",
        config.base_url
    ))
}

#[post("/")]
async fn paste(
    req: web::HttpRequest,
    mut body: web::Payload,
    config: web::Data<AppConfig>,
) -> impl Responder {
    let key = crate::bip39::mnemonic(config.key_size);
    let content_length: Option<i64> = req
        .headers()
        .get(header::CONTENT_LENGTH)
        .and_then(|x| x.to_str().ok())
        .and_then(|x| x.parse().ok());
    let (tx, rx) = tokio::sync::mpsc::unbounded_channel();
    rt::spawn(async move {
        while let Some(chunk) = body.next().await {
            match chunk {
                Ok(bytes) => tx.send(Ok(bytes)),
                Err(e) => tx.send(Err(Error::new(ErrorKind::Other, e))),
            }
            .unwrap()
        }
    });
    let req = PutObjectRequest {
        bucket: config.bucket.clone(),
        key: key.clone(),
        body: Some(ByteStream::new(UnboundedReceiverStream::new(rx))),
        content_length,
        ..PutObjectRequest::default()
    };
    let resp = config.client.put_object(req).await;
    match resp {
        Ok(_) => HttpResponse::Ok().body(format!(
            "{}\n",
            config.base_url.join(&key).unwrap().to_string()
        )),
        Err(_) => HttpResponse::InternalServerError().finish(),
    }
}

#[get("/{id}")]
async fn retrieve(id: web::Path<String>, config: web::Data<AppConfig>) -> impl Responder {
    let req = GetObjectRequest {
        bucket: config.bucket.clone(),
        key: id.to_string(),
        ..GetObjectRequest::default()
    };
    let resp = config.client.get_object(req).await;
    match resp {
        Ok(resp) => match resp.body {
            Some(body) => {
                let mut ret = HttpResponse::Ok();
                if let Some(content_length) = resp.content_length {
                    ret.no_chunking(content_length as u64);
                }
                ret.streaming(body)
            }
            None => HttpResponse::NoContent().finish(),
        },
        Err(_) => HttpResponse::NotFound().finish(),
    }
}

struct AppConfig {
    base_url: Url,
    key_size: usize,
    client: S3Client,
    bucket: String,
}

async fn _main() -> Result<()> {
    let port = env::var("PORT").unwrap_or_else(|_| String::from("8080"));
    let base_url = Url::parse(&env::var("BASE_URL")?)?;
    let endpoint = env::var("S3_ENDPOINT")?;
    let bucket = env::var("S3_BUCKET")?;
    let region = env::var("S3_REGION")?;
    let access_key = env::var("S3_ACCESS_KEY_ID")?;
    let secret_key = env::var("S3_SECRET_ACCESS_KEY")?;
    let region = Region::Custom {
        name: region,
        endpoint,
    };
    let creds = StaticProvider::new_minimal(access_key, secret_key);
    let hclient = HttpClient::new()?;
    let client = S3Client::new_with(hclient, creds, region);
    HttpServer::new(move || {
        App::new()
            .data(AppConfig {
                base_url: base_url.clone(),
                key_size: 3,
                client: client.clone(),
                bucket: bucket.clone(),
            })
            .service(index)
            .service(paste)
            .service(retrieve)
    })
    .bind(format!("[::]:{}", port))?
    .run()
    .await?;
    Ok(())
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    _main()
        .await
        .map_err(|e| std::io::Error::new(std::io::ErrorKind::Other, e))
}
