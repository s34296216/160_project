mod file_reader;
use file_reader::{read_file, Inventory};
mod watcher;
use watcher::file_watcher;
mod categories;
use categories::inventory::inv_filtering ;
//use serde::{Deserialize, Serialize};
use tokio;
use async_nats::connect;

/* 
#[derive(Debug, Serialize, Deserialize)]
struct StockDetail
{ 
    stock: u32,
    change: Option<i32>,
    reason: Option<String>
}

#[derive(Debug, Serialize, Deserialize)]
struct InventoryUpdate {
    timestamp: String,
    product_id: u32,
    event: String,
    details: StockDetail,
    status: String,
    message: String
}


fn inv_filtering(inventories: &Vec<Inventory>) -> Result<String, serde_json::Error> {
    let mut filtered = Vec::new();
    
    for inventory in inventories.iter() {
        
        let current_stock = inventory.stock;
        let mut status = String::new();
        let mut message = String::new();
        if current_stock >= 50
        {
            status = "ok".to_string();
            message = "Inventory updated successfully.".to_string();
        }
        else if  current_stock >=10  
        {
             status = "reminder".to_string();
             message = "Low inventory.".to_string();
        }
        else if current_stock < 10
        {
             status = "warning".to_string();
             message = "Immediate restocking required.".to_string();
        }
        
        let update = InventoryUpdate {
            timestamp: inventory.timestamp.clone(),
            product_id: inventory.product_id,
            event: inventory.event.clone(),
            details: StockDetail {
                stock: inventory.stock,
                change: inventory.change,
                reason: inventory.reason.clone()
            },
            status,
            message
        };
        filtered.push(update);
    }
    
    return serde_json::to_string_pretty(&filtered);
}
*/

/* 
fn display(vec: &Vec<file_reader::Inventory>) {
    println!("The processed logs are:");
    for inventory in vec {
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
*/



async fn publish_message(filtered_data: String, topic: String) -> Result<(), Box<dyn std::error::Error + Send + Sync >> {
    let client = connect("nats://localhost:4222").await?;
    println!("Connected to NATS server");
    
    client.publish(topic, filtered_data.into()).await?;
    client.flush().await?;
    
    Ok(())
}




#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let file_name = "test_log.xml".to_string();
    let mut position = 0;
    let inventories = read_file(&file_name, &mut position).await?;
    println!("current entry: {}", &position);
  //  display(&inventories);
    
    let filtered_data = inv_filtering(&inventories)?;
    println!("Filtered data:\n{}", filtered_data);
    
    publish_message(filtered_data, "inventory.updates".to_string()).await?;


    // let mut stock: Vec<u32> = Vec::new();
    // for inventory in &inventories{
    //     stock.push(inventory.stock);
    // }
    // for i in 0..stock.len(){
    //     println!("{}", stock[i]);
    // }


   /* 
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
    */
  //  let index :usize = position;
        let mut position2 = position.saturating_add(1);
    loop {
        if file_watcher(&file_name).await? {
            println!("File modified");
            
          //  let mut inventories:Vec<Inventory> = Vec::new();
          //  let mut new_index :usize = 0;
          // let (inventories, posistion) = read_file(&file_name, position.saturating_add(1)).await?;
           let inventories = read_file(&file_name, &mut position2).await?;
         //   println!("current entry: {}", &posistion2);
         //   display(&inventories);
         //   position = index;
         //   break;
             let new_data = inv_filtering(&inventories)?;
             println!("Filtered data:\n{}", new_data);
           
             publish_message(new_data, "inventory.updates".to_string()).await?;
        }
        tokio::time::sleep(tokio::time::Duration::from_millis(100)).await;
    }
    
    Ok(()) // like return 0 in c++
}