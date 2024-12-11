
use std::{sync::mpsc::channel, time::{Duration, Instant}, path::Path, fs};
use notify::{Config, RecommendedWatcher, RecursiveMode, Watcher};

async fn notice(delay:Duration, last_time:&mut Instant) ->bool
{
    let current_time = Instant::now();
    if current_time.duration_since(*last_time) > delay{
         println!("file change");
         *last_time = current_time;

        return true;

    }
    return false;
}



pub async fn file_watcher(file_names: &Vec<&str>, inv_pos:&mut usize, pag_pos:&mut usize) -> Result< bool, Box<dyn std::error::Error + Send + Sync>> {
    let (tx, rx) = channel();

    let mut watcher = RecommendedWatcher::new(tx, Config::default())?;
    //let mut last_position = 0;
    let delay = Duration::from_millis(250);
    
    let path_1 =Path::new(file_names[0]);
    let path_2 = Path::new(file_names[1]);

    watcher.watch(path_1, RecursiveMode::Recursive)?;
    watcher.watch(path_2, RecursiveMode::Recursive)?;
    

    println!("Start watching");
    let mut last_time = Instant::now();
    for res in rx {
        if let Ok(event) = res {
            
            if let notify::EventKind::Modify(_) = event.kind {
                
                println!("I am here");
                for path in event.paths {
                    
                    let event_path = fs::canonicalize(&path)?;
                    let abs_path1 = fs::canonicalize(path_1)?;
                    let abs_path2 = fs::canonicalize(path_2)?;
                   
                    if event_path.to_str()==abs_path1.to_str() {
                        println!("matched inventory logs");
                        if notice(delay, &mut last_time).await {
                            
                            let mut new_inv_pos = *inv_pos;
                            let inventories = crate::file_reader::read_inventory_logs(&file_names[0], &mut new_inv_pos).await?;
                            let filtered_inv = crate::inv_filtering(&inventories)?;
                            println!("{}", filtered_inv);
                            crate::publish_message(filtered_inv, "inventory.updates".to_string()).await?;
                            *inv_pos = new_inv_pos;

                            return Ok(true);
                        }
                    }
                    else if event_path.to_str()==abs_path2.to_str() {
                        println!("matched page logs");
                        if notice(delay, &mut last_time).await {
                            
                            let mut new_pag_pos = *pag_pos;
                            let pag_logs = crate::file_reader::read_page_logs(&file_names[1], &mut new_pag_pos).await?;
                            let filtered_logs = crate::pag_filtering(&pag_logs)?;
                            println!("{}", filtered_logs);
                           // crate::publish_message(filtered_logs, "page.response".to_string()).await?;
                            *pag_pos = new_pag_pos;
                            
                            
                            return Ok(true);
                        }
                    }
                    else {
                        println!("not matched");

                    }
                }
                
                
                // if notice(delay, &mut last_time).await {
                //     return Ok(true);
                // }
            }
        }
    }

    Ok(false)
}