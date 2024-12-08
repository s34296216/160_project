//Wait for multithreading

use quick_xml::de::from_reader;
use serde::Deserialize;
use tokio::fs::File;
use tokio::io::{BufReader, AsyncReadExt};

use std::error::Error;
//use std::string;


#[derive(Debug, Deserialize)]
struct Logs { 
    #[serde(rename = "log")]
    logs: Vec<Inventory>, 
}

/* 
#[derive(Debug, Deserialize, Clone)] //need to be pubic for struct and its members
pub struct Inventory { 
    pub timestamp: String,
    pub event: String,
    pub product_id: u32,
    pub stock: u32,
    #[serde(default)]
    pub change: Option<i32>,
    #[serde(default)]
    pub reason: Option<String>,
    #[serde(default)]
    pub action: Option<String>,
}
*/

#[derive(Debug, Deserialize, Clone, Default)]
pub struct Inventory { 
    #[serde(default)]
    pub timestamp: String,
    #[serde(default)]
    pub event: String,
    #[serde(default)]
    pub product_id: u32,
    #[serde(default)]
    pub stock: u32,
    #[serde(default)]
    pub change: Option<i32>,
    #[serde(default)]
    pub reason: Option<String>,
    #[serde(default)]
    pub action: Option<String>,
}

pub async fn read_file(file_name: &str, start: &mut usize) -> Result< Vec<Inventory>,  Box<dyn Error + Send + Sync>> { // uniform error detection
    let file = File::open(file_name).await?;
    
    // let start_position = 
    // if let Some(pos) = start {
    //     pos.saturating_sub(1)
    // } else {
    //     0
    // };

    let start_position = start.saturating_sub(1);
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
 //   Ok(logs.logs)
   // let currect_position = start_position;
//    let partial_logs = 
//    if start_position >= logs.logs.len() {
//     Vec::new()
//    } else {
//     logs.logs[start_position..].to_vec()
//    };

    println!("The entire len is {}", logs.logs.len());
  //  let index = logs.logs.len();
    
    let partial_logs = if start_position == 0 {
        logs.logs.clone()
    } else if start_position < logs.logs.len() {
        vec![logs.logs.last().cloned().unwrap_or_default()]
    } else {
        Vec::new()
    };
    
    println!("The len is {}", partial_logs.len());
    // let mut current_position = 0;
    // current_position = start_position + partial_logs.len();
    *start = start_position + partial_logs.len();

    
    Ok(partial_logs)
}

/* 
#[tokio::main]
async fn main() {
    let file_name = "test_log.xml".to_string();
    read_file(&file_name).await.unwrap();
}
*/