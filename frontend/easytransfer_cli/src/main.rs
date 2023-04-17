use std::{
    env,
    io::{self, stdout, Write},
    process::exit, thread, time::Duration,
};

use reqwest::blocking::Client;
use serde::{Deserialize, Serialize};

const API_HEALTH: &str = "http://localhost:3287/health";
const API_DEVICES: &str = "http://localhost:3287/devices";
const API_DEVICES_CACHE: &str = "http://localhost:3287/devices_cache";

#[derive(Debug, Serialize, Deserialize, Clone)]
struct Device {
    ip: String,
    name: String,
}

fn refresh_devices(client: &Client, devices: &mut Vec<Device>) {
    println!("Getting updated list of devices...");
    *devices = client
        .get(API_DEVICES)
        .send()
        .unwrap()
        .json::<Vec<Device>>()
        .unwrap();
}

fn is_device_cache_stale(client: &Client, devices: &Vec<Device>) -> bool {
    println!("Checking device list cache for staleness...");

    for d in devices {
        let name = client
            .get(format!("http://{}:3287/device/name", d.ip))
            .send()
            .unwrap()
            .text()
            .unwrap();
        if name != d.name {
            return true;
        }
    }
    false
}

fn exit_error_bg_error() {
    exit_error("Could not connect to the local EasyTransfer server on port 3287!");
}

fn exit_error(msg: &str) {
    eprintln!("Error: {}", msg);
    exit(1);
}

fn send_file(client: &Client, file: &str, ip: &str) {
    let form = reqwest::blocking::multipart::Form::new()
        .file("file", file);

    if form.is_err() {
       exit_error(&format!("File: {} does not exist", file));
    }

    let res = client
        .post(format!("http://{}:3287/upload", ip))
        .multipart(form.unwrap())
        .timeout(Duration::from_secs(120))
        .send()
        .unwrap()
        .text()
        .expect("Error sending file");

    if res != "Ok" {
        exit_error("Could not upload file");
    }

    println!("File {} sent!", file);
}

fn server_health_check(client: &Client) {
    println!(
        "
███████  █████  ███████ ██    ██ ████████ ██████   █████  ███    ██ ███████ ███████ ███████ ██████  
██      ██   ██ ██       ██  ██     ██    ██   ██ ██   ██ ████   ██ ██      ██      ██      ██   ██ 
█████   ███████ ███████   ████      ██    ██████  ███████ ██ ██  ██ ███████ █████   █████   ██████  
██      ██   ██      ██    ██       ██    ██   ██ ██   ██ ██  ██ ██      ██ ██      ██      ██   ██ 
███████ ██   ██ ███████    ██       ██    ██   ██ ██   ██ ██   ████ ███████ ██      ███████ ██   ██ 
                                                                                                    
                                                                                                    
"
    );
    let resp = client.get(API_HEALTH).send();

    if resp.is_err() {
        exit_error_bg_error();
    }

    if resp.unwrap().text().unwrap() != "Healthy!" {
        exit_error_bg_error();
    }

    println!("Local background server is online");
}

fn get_devices(client: &Client) -> Vec<Device> {
    println!("Getting list of devices...");
    let devices: Vec<Device> = client
        .get(API_DEVICES_CACHE)
        .send()
        .unwrap()
        .json::<Vec<Device>>()
        .unwrap();
    devices
}

fn device_selection(devices: &mut Vec<Device>) -> usize {
    if devices.len() == 0 {
        exit_error("No devices connected to your local network");
    }

    println!("-----------------------");
    let mut i = 0;
    for d in devices.clone() {
        i += 1;
        println!("{} - {} ({})", i, d.name, d.ip);
    }

    println!("-----------------------");
    let mut selected_device_idx;
    print!("Select device to send the files to: ");
    let _ = stdout().flush();
    loop {
        let mut line = String::new();
        io::stdin()
            .read_line(&mut line)
            .expect("Failed to read input");
        let device_idx = line.trim().parse::<usize>();
        match device_idx {
            Ok(i) => {
                selected_device_idx = i - 1;
                match devices.get(selected_device_idx) {
                    Some(_) => return selected_device_idx,
                    None => print!("Please select an existing device number: "),
                }
            }
            Err(_) => print!("Please insert a number: "),
        }
        let _ = stdout().flush();
    }
}

fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() <= 1 {
        println!("Usage: easytransfer <file1 file2 file3 ...>");
        exit(0);
    }

    let client = reqwest::blocking::Client::new();

    // If bg server is not running on client, abort
    server_health_check(&client);

    // Get cached devices from server
    let mut devices = get_devices(&client);

    let mut selected_device_idx = device_selection(&mut devices);

    if is_device_cache_stale(&client, &devices) {
        refresh_devices(&client, &mut devices);
        selected_device_idx = device_selection(&mut devices);
    }

    let mut threads = vec![];
    for a in args.into_iter().skip(1) {
        let d = devices.get(selected_device_idx).unwrap().clone();
        println!(
            "Sending {} to {} ...",
            a,
            d.name
        );
        threads.push(thread::spawn(move || {
            send_file(&reqwest::blocking::Client::new(), &a, &d.ip);
         }));
    }

    for t in threads {
        let _ = t.join();
    }
}
