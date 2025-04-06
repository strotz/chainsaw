fn main() -> anyhow::Result<()> {
    println!("Hello, world!");
    println!("{}", uuid::Uuid::new_v4());
    Ok(())
}
