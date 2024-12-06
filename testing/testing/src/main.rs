mod file_reader;
use file_reader::read_file;

use tokio;


#[tokio::main]
async fn main() {
    let file_name = "test_log.xml".to_string();
    let inventories = read_file(&file_name).await.unwrap();

    for inventory in &inventories { //use reference becasue no need to own
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

}