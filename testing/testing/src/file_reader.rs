use quick_xml::de::from_reader;
use serde::Deserialize;
use tokio::fs::File;
use tokio::io::{BufReader, AsyncReadExt};
use std::error::Error;

#[derive(Debug, Deserialize)]
struct PageLogs { 
    #[serde(rename = "log")]
    logs: Vec<PageResponse>, 
}

#[derive(Debug, Deserialize)]
struct InventoryLogs { 
    #[serde(rename = "log")]
    logs: Vec<Inventory>, 
}

#[derive(Debug, Deserialize, Clone, Default)]
pub struct PageResponse {
    pub timestamp: String,
    pub event: String,
    pub user_ip: String,
    pub endpoint: String,
    pub response_time_ms: u32,
}

#[derive(Debug, Deserialize, Clone, Default)]
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

pub async fn read_page_logs(file_name: &str, start: &mut usize) -> Result<Vec<PageResponse>, Box<dyn Error + Send + Sync>> {
    let file = File::open(file_name).await?;
    let start_position = start.saturating_sub(1);
    let mut reader = BufReader::new(file);
    
    let mut contents = String::new();
    reader.read_to_string(&mut contents).await?;
    println!("Raw XML: {}", contents);
    
    let logs: PageLogs = from_reader(contents.as_bytes())?;
    
    println!("The entire len is {}", logs.logs.len());
    
    let partial_logs = if start_position == 0 {
        logs.logs.clone()
    } else if start_position < logs.logs.len() {
        vec![logs.logs.last().cloned().unwrap_or_default()]
    } else {
        Vec::new()
    };
    
    println!("The len is {}", partial_logs.len());
    *start = start_position + partial_logs.len();
    
    Ok(partial_logs)
}

pub async fn read_inventory_logs(file_name: &str, start: &mut usize) -> Result<Vec<Inventory>, Box<dyn Error + Send + Sync>> {
    let file = File::open(file_name).await?;
    let start_position = start.saturating_sub(1);
    let mut reader = BufReader::new(file);
    
    let mut contents = String::new();
    reader.read_to_string(&mut contents).await?;
    println!("Raw XML: {}", contents);
    
    let logs: InventoryLogs = from_reader(contents.as_bytes())?;
    
    println!("The entire len is {}", logs.logs.len());
    
    let partial_logs = if start_position == 0 {
        logs.logs.clone()
    } else if start_position < logs.logs.len() {
        vec![logs.logs.last().cloned().unwrap_or_default()]
    } else {
        Vec::new()
    };
    
    println!("The len is {}", partial_logs.len());
    *start = start_position + partial_logs.len();
    
    Ok(partial_logs)
}