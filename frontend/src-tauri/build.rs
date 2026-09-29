fn main() {
    // 16KB ページの端末でも読み込めるようにする（Google Play の要件）。Tauri の CLI が RUSTFLAGS を上書きするためここで渡す
    if std::env::var("CARGO_CFG_TARGET_OS").as_deref() == Ok("android") {
        println!("cargo:rustc-link-arg=-Wl,-z,max-page-size=16384");
    }
    tauri_build::build()
}
