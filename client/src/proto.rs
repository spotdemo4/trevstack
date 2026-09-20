pub mod auth {
    #[allow(clippy::empty_docs)]
    pub mod v1 {
        include!("connect/auth.v1.rs");
    }
}

pub mod number {
    #[allow(clippy::empty_docs)]
    pub mod v1 {
        include!("connect/number.v1.rs");
    }
}
