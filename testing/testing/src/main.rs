mod file_reader;
use file_reader::read_file;
mod watcher;
use watcher::file_watcher;

use tokio;


#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let file_name = "test_log.xml".to_string();
    let (inventories, position) = read_file(&file_name, None).await?; //default 1, or use Some(N), N>=1
    println!("current entry: {}", position );
    
    for inventory in &inventories { // use reference becasue no n eed to own
        println!("inventory:");
        println!("  Timestamp: {}", inventory.timestamp);
        println!("  Event: {}", inventory.event);
        println!("  Product ID: {}", inventory.product_id);
        println!("  Stock: {}", inventory.stock);
        
        if let Some(change) = inventory.change {
            println!("  Change: {}", change);
        }
        
        if let Some(reason) = &inventory.reason {
            println!("  Reason: {}", reason);
        }
        
        if let Some(action) = &inventory.action {
            println!("  Action: {}", action);
        }
        println!();
    }
    
    loop {
        if file_watcher(&file_name).await? {
            println!("File modified");
            break;
        }
        tokio::time::sleep(tokio::time::Duration::from_millis(100)).await;
    }
    Ok(()) // like return 0 in c++
}