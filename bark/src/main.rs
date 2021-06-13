use lettre::transport::smtp::{authentication::Credentials, AsyncSmtpTransport};
use lettre::{AsyncTransport, Message};
use rocket::form::{Form, FromForm};
use rocket::http::Status;
use serde::Deserialize;

#[rocket::main]
async fn main() {
    let rocket = rocket::build();
    let config: Config = rocket.figment().extract().unwrap();
    let rocket = rocket.manage(config);
    let rocket = rocket.mount("/", rocket::routes![webhook]);
    rocket.launch().await.unwrap();
}

#[derive(Deserialize)]
struct Config {
    smtp_server: String,
    smtp_username: String,
    smtp_password: String,
    mail_address: String,
}

#[derive(FromForm)]
struct SMS<'a> {
    #[field(name = "From")]
    from: &'a str,
    #[field(name = "To")]
    to: &'a str,
    #[field(name = "Body")]
    body: &'a str,
}

fn err<T: std::error::Error>(e: T) -> anyhow::Error {
    anyhow::anyhow!(e.to_string())
}

#[rocket::post("/webhook", data = "<sms>")]
async fn webhook(
    config: &rocket::State<Config>,
    sms: Form<SMS<'_>>,
) -> Result<Status, rocket::response::Debug<anyhow::Error>> {
    let email = Message::builder()
        .from(config.mail_address.parse().unwrap())
        .to(config.mail_address.parse().unwrap())
        .subject(format!("Received SMS from {} to {}", sms.from, sms.to))
        .body(String::from(sms.body))
        .map_err(err)?;
    let sender = AsyncSmtpTransport::<lettre::Tokio1Executor>::relay(&config.smtp_server)
        .map_err(err)?
        .credentials(Credentials::new(
            config.smtp_username.clone(),
            config.smtp_password.clone(),
        ))
        .build();
    let result = sender.send(email).await.map_err(err)?;
    if !result.is_positive() {
        Err(anyhow::anyhow!(result.code()))?;
    }
    Ok(Status::Ok)
}
