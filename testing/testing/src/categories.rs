pub mod inventory //like scope in c++
{
    use serde::{Deserialize, Serialize};
    use crate::file_reader::Inventory;

    #[derive(Debug, Serialize, Deserialize)]
    pub struct StockDetail
    {
    pub stock: u32,
    pub change: Option<i32>,
    pub reason: Option<String>
    }

    #[derive(Debug, Serialize, Deserialize)]
    pub struct InventoryUpdate {
        pub timestamp: String,
        pub product_id: u32,
        pub event: String,
        pub details: StockDetail,
        pub status: String,
        pub message: String
    }

    pub fn inv_filtering(inventories: &Vec<Inventory>) -> Result<String, serde_json::Error> {
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

}


pub mod page_response {
    use serde::{Deserialize, Serialize};
    use crate::file_reader::PageResponse;

    #[derive(Debug, Serialize, Deserialize)]
    pub struct PageResponseUpdate {
        pub timestamp: String,
        pub user_ip: String,
        pub event: String,
        pub endpoint: String,
        pub response_time_ms: String,
        pub status: String,
    }

    pub fn pag_filtering(responses: &Vec<PageResponse>) -> Result<String, serde_json::Error> {
        let mut filtered = Vec::new();
        
        for response in responses.iter() {
            
            let response_time = response.response_time_ms;
            let mut status = String::new();

            if response_time > 1000 {
                status = "slow".to_string();
            }
            else {
                status = "normal".to_string();
            }
            
            let update = PageResponseUpdate {
                timestamp: response.timestamp.clone(),
                user_ip: response.user_ip.clone(),
                event: response.event.clone(),
                endpoint: response.endpoint.clone(),
                response_time_ms: response.response_time_ms.to_string(),
                status,
            };
            filtered.push(update);
        }
        
        serde_json::to_string_pretty(&filtered)
    }
}