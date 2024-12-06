//Wait for multithreading

use quick_xml::de::from_reader;
use serde::Deserialize;
use tokio::fs::File;
use tokio::io::{BufReader, AsyncReadExt};
//use tokio::io::AsyncReadExt;
use std::error::Error;
//use std::string;


#[derive(Debug, Deserialize)]
struct Logs { 
    #[serde(rename = "log")]
    logs: Vec<Inventory>, 
}

#[derive(Debug, Deserialize)] //need to be pubic for struct and its members
pub struct Inventory { 
    pub timestamp: String,
    pub event: String,
    pub product_id: u32,
    pub stock: i32,
    #[serde(default)]
    pub change: Option<i32>,
    #[serde(default)]
    pub reason: Option<String>,
    #[serde(default)]
    pub action: Option<String>,
}


pub async fn read_file(file_name: &str) -> Result<Vec<Inventory>, Box<dyn Error>> { // uniform error detection
    let file = File::open(file_name).await?;
    let mut reader = BufReader::new(file);
    
 
    let mut contents = String::new();
    reader.read_to_string(&mut contents).await?;
    
    
    let logs: Logs = from_reader(contents.as_bytes())?; // parse the XML string
    
    /*   
    for log in &logs.logs {
        println!("inventory:");
        println!("  Timestamp: {}", log.timestamp);
        println!("  Event: {}", log.event);
        println!("  Product ID: {}", log.product_id);
        println!("  Stock: {}", log.stock);
        
        if let Some(change) = log.change {
            println!("  Change: {}", change);
        }
        
        if let Some(reason) = &log.reason {
            println!("  Reason: {}", reason);
        }
        
        if let Some(action) = &log.action {
            println!("  Action: {}", action);
        }
        println!();
    }
    */
    Ok(logs.logs)
}

/* 
#[tokio::main]
async fn main() {
    let file_name = "test_log.xml".to_string();
    read_file(&file_name).await.unwrap();
}
*/