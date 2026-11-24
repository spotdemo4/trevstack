import org.gradle.api.artifacts.dsl.LockMode

plugins {
    alias(libs.plugins.android.application) apply false
    alias(libs.plugins.kotlin.compose) apply false
}

// The app's version name, which bumper updates. Its version code is derived from it.
version = "0.2.1"

// every resolved version is pinned in gradle.lockfile; nix run .#update-android-deps rewrites it
allprojects {
    dependencyLocking {
        lockAllConfigurations()
        lockMode = LockMode.STRICT
    }
}
