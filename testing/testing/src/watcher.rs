
use std::{sync::mpsc::channel, time::{Duration, Instant}, path::Path};
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



pub async fn file_watcher(file_name: &str) -> Result< bool, Box<dyn std::error::Error + Send + Sync>> {
    let (tx, rx) = channel();

    let mut watcher = RecommendedWatcher::new(tx, Config::default())?;
    //let mut last_position = 0;
    let delay = Duration::from_millis(250);

    watcher.watch(Path::new(file_name), RecursiveMode::Recursive)?;
    println!("Start watching");
    let mut last_time = Instant::now();
    for res in rx {
        if let Ok(event) = res {
            
            if let notify::EventKind::Modify(_) = event.kind {
                if notice(delay, &mut last_time).await {
                    return Ok(true);
                }
            }
        }
    }

    Ok(false)
}