pub mod auth {
    #[allow(clippy::empty_docs)]
    pub mod v1 {
        include!("gen/auth.v1.rs");
        include!("gen/auth.v1.trevrpc.rs");
    }
}

pub mod number {
    #[allow(clippy::empty_docs)]
    pub mod v1 {
        include!("gen/number.v1.rs");
        include!("gen/number.v1.trevrpc.rs");
    }
}
