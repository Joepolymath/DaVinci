use shared_rust::utils::sysinfo;

fn main() {
    println!("DaVinci shared-rust v{}", shared_rust::version());
    println!("CPU cores: {}", sysinfo::cpu_count());
    println!("Unix timestamp: {}", sysinfo::unix_timestamp());
}

