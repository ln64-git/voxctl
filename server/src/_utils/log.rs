// region: --- imports
use chrono::prelude::*; // Import everything from the prelude module
use std::fs::OpenOptions;
use std::io::Write;
// endregion: --- imports

pub fn custom_log(message: &str) {
    let timestamp = Local::now().format("%Y-%m-%d %H:%M:%S");
    let log_entry = format!("{} - {}\n", timestamp, message);

    println!("{}", log_entry); // Print the log entry to the console

    // Write the log entry to the log file
    let mut file = OpenOptions::new()
        .append(true)
        .create(true)
        .open(format!(
            "server/logs/{}.log",
            Local::now().format("%Y-%m-%d")
        ))
        .unwrap();
    file.write_all(log_entry.as_bytes()).unwrap();
}
