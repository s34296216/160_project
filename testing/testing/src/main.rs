mod file_reader;
use std::vec;

use file_reader::{read_page_logs, read_inventory_logs };
mod watcher;
use watcher::file_watcher;
mod categories;
use categories::inventory::inv_filtering ;
use categories::page_response::pag_filtering;
//use serde::{Deserialize, Serialize};
use tokio::task;
use async_nats::connect;





async fn display_inv(vec: &Vec<file_reader::Inventory>) {
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

        // if let Some(action) = &inventory.action {
        //     println!("  Action: {}", action);
        // }
        println!();
    }
}  

async fn display_pag(vec: &Vec<file_reader::PageResponse>) {
    println!("The processed logs are:");
    for page in vec {
        println!("page_response:");
        println!("  Timestamp: {}", page.timestamp);
        println!("  Event: {}", page.event);
        println!("  User IP: {}", page.user_ip);
        println!("  Endpoint: {}", page.endpoint);
        println!("  Response Time: {}ms", page.response_time_ms);
        println!();
    }
}



async fn publish_message(filtered_data: String, topic: String) -> Result<(), Box<dyn std::error::Error + Send + Sync >> {
    let client = connect("nats://localhost:4222").await?;
    println!("Connected to NATS server");
    
    client.publish(topic, filtered_data.into()).await?;
    client.flush().await?;
    
    Ok(())
}



#[tokio::main(flavor = "multi_thread", worker_threads = 2)] 
//#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let file_name = "test_log2.xml".to_string();
    let file_name_2 = "test_log.xml".to_string();

    let paths=vec!["test_log.xml", "test_log2.xml" ];
    // if file_watcher(paths).await?
    // {
    //     println!("succeeded");
    // }

      
    let task_1 = task::spawn(async move{
        let mut position = 0;
        let inventories = read_inventory_logs(&file_name_2, &mut position).await?;
        display_inv(&inventories).await;
        let filtered_data = inv_filtering(&inventories)?;
        println!("{}", filtered_data);
        publish_message(filtered_data, "inventory.updates".to_string()).await?;
        Ok::<usize, Box<dyn std::error::Error + Send + Sync>>(position)
    });

    let task_2 = task::spawn(async move{
        let mut position = 0;
        let page_logs = read_page_logs(&file_name, &mut position).await?;
        display_pag(&page_logs).await;
        let filtered_data = pag_filtering(&page_logs)?;
        println!("{}", filtered_data);
     //   publish_message(filtered_data, "page.response".to_string()).await?;
        Ok::<usize, Box<dyn std::error::Error + Send + Sync>>(position)
    });
    let (inv_pos,pag_pos) = tokio::join!(task_1, task_2);




   // let mut position = 0;
   // let inventories = read_file(&file_name, &mut position).await?;
   // let page_logs = read_page_logs(&file_name, &mut position).await?;
   // println!("current entry: {}", &position);
  //  display(&inventories);
   // display_pag(&page_logs);


   


    
   //  let filtered_data = pag_filtering(&page_logs)?;
   //  println!("Filtered data:\n{}", filtered_data);
    
    // publish_message(filtered_data, "inventory.updates".to_string()).await?;


    // let mut stock: Vec<u32> = Vec::new();
    // for inventory in &inventories{
    //     stock.push(inventory.stock);
    // }
    // for i in 0..stock.len(){
    //     println!("{}", stock[i]);
    // }

    
        let inv_pos = inv_pos??; // Unwrap the Result and handle the error if it occurs
        let pag_pos = pag_pos??;
 
    
        let mut pag_pos2 = pag_pos.saturating_add(1); 
        let mut inv_pos2 = inv_pos.saturating_add(1);

        println!("Into the loop");
    loop {
        if file_watcher(&paths, &mut inv_pos2, &mut pag_pos2 ).await? {
            println!("File modified");
            
        //    let mut inventories:Vec<Inventory> = Vec::new();
        //    let mut new_index :usize = 0;
        //   let (inventories, posistion) = read_file(&file_name, position.saturating_add(1)).await?;
        //   let logs = read_page_logs(&file_name, &mut pag_pos2).await?;
        //    println!("current entry: {}", &posistion2);
        //    display(&inventories);
        //    position = index;
        //    break;
            //  let new_data = pag_filtering(&logs)?;
            //  println!("Filtered data:\n{}", new_data);
           
           //  publish_message(new_data, "inventory.updates".to_string()).await?;
        }
        tokio::time::sleep(tokio::time::Duration::from_millis(100)).await;
    }
    
    #[allow(unreachable_code)]
    Ok(()) // like return 0 in c++
}