import org.gradle.api.DefaultTask
import org.gradle.api.file.ConfigurableFileCollection
import org.gradle.api.file.DirectoryProperty
import org.gradle.api.tasks.InputFiles
import org.gradle.api.tasks.OutputDirectory
import org.gradle.api.tasks.PathSensitive
import org.gradle.api.tasks.PathSensitivity
import org.gradle.api.tasks.TaskAction

plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.compose)
}

/** Semantic version parts of the root project's version, such as 1.2.3. */
val versionParts: List<Int> =
    rootProject.version
        .toString()
        .split('.')
        .map { part -> requireNotNull(part.toIntOrNull()) { "version must be MAJOR.MINOR.PATCH, not ${rootProject.version}" } }
        .also { parts -> require(parts.size == 3 && parts.all { it in 0..999 }) { "version must be MAJOR.MINOR.PATCH, each 0 to 999" } }

android {
    namespace = "zip.trev.template.stack"
    compileSdk {
        version = release(37) { minorApiLevel = 2 }
    }
    buildToolsVersion = "37.0.0"

    defaultConfig {
        applicationId = "zip.trev.template.stack"
        // Cronet 500 needs Android 7.0
        minSdk = 24
        targetSdk = 37
        versionCode = versionParts[0] * 1_000_000 + versionParts[1] * 1_000 + versionParts[2]
        versionName = rootProject.version.toString()
        // the server the sign-in screen starts with; pass -PstackUrl=https://example.com to change it
        buildConfigField(
            "String",
            "DEFAULT_URL",
            "\"${providers.gradleProperty("stackUrl").getOrElse("https://localhost:8080")}\"",
        )
    }

    buildTypes {
        release {
            // R8 fails on TrevRPC core's references to zstd-jni, which it compiles against but
            // doesn't ship, and has no rules for
            isMinifyEnabled = false
        }
    }

    buildFeatures {
        buildConfig = true
        compose = true
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    // the messages and clients that buf generates from ../proto
    sourceSets.getByName("main") {
        java.directories.add("src/gen")
        kotlin.directories.add("src/gen")
    }

    lint {
        warningsAsErrors = true
        // newer dependencies arrive through Renovate, which lint can't see
        disable += setOf("AndroidGradlePluginVersion", "GradleDependency", "NewerVersionAvailable")
    }
}

/**
 * Writes the debug build's network security configuration, which trusts the development
 * certificate of a server on this machine as well as the system's certificates. The server writes
 * the certificate to development.pem in the trevstack directory of the user configuration
 * directory, or the file is given with -PstackCaCert=FILE. Without it, debug builds trust only the
 * system's certificates, as release builds do.
 */
abstract class DevelopmentCertificateConfig : DefaultTask() {
    /** The certificate file, if it exists. */
    @get:InputFiles
    @get:PathSensitive(PathSensitivity.NONE)
    abstract val certificate: ConfigurableFileCollection

    @get:OutputDirectory
    abstract val outputDirectory: DirectoryProperty

    @TaskAction
    fun write() {
        val output = outputDirectory.get().asFile
        output.deleteRecursively()
        val pem = certificate.files.firstOrNull { it.isFile }
        val anchors =
            if (pem == null) {
                ""
            } else {
                output.resolve("raw").mkdirs()
                pem.copyTo(output.resolve("raw/development_certificate.pem"))
                """
                |    <debug-overrides>
                |        <trust-anchors>
                |            <certificates src="system" />
                |            <certificates src="@raw/development_certificate" />
                |        </trust-anchors>
                |    </debug-overrides>
                |
                """.trimMargin()
            }
        output.resolve("xml").mkdirs()
        output.resolve("xml/network_security_config.xml").writeText(
            """
            |<?xml version="1.0" encoding="utf-8"?>
            |<network-security-config>
            |    <base-config cleartextTrafficPermitted="false" />
            |$anchors</network-security-config>
            |
            """.trimMargin(),
        )
    }
}

/** The user configuration directory, as the server's os.UserConfigDir finds it. */
fun userConfigDirectory(): File? {
    val home = providers.systemProperty("user.home").orNull
    val os = providers.systemProperty("os.name").get().lowercase()
    return when {
        os.startsWith("windows") -> {
            providers.environmentVariable("APPDATA").orNull?.let(::File)
        }

        os.startsWith("mac") -> {
            home?.let { File(it, "Library/Application Support") }
        }

        else -> {
            providers.environmentVariable("XDG_CONFIG_HOME").orNull?.let(::File)
                ?: home?.let { File(it, ".config") }
        }
    }
}

val developmentCertificateConfig =
    tasks.register<DevelopmentCertificateConfig>("developmentCertificateConfig") {
        val configured = providers.gradleProperty("stackCaCert").orNull?.let(::File)
        certificate.from(configured ?: userConfigDirectory()?.resolve("trevstack/development.pem") ?: emptyList<File>())
        outputDirectory.set(layout.buildDirectory.dir("generated/developmentCertificate"))
    }

androidComponents {
    onVariants(selector().withBuildType("debug")) { variant ->
        variant.sources.res?.addGeneratedSourceDirectory(
            developmentCertificateConfig,
            DevelopmentCertificateConfig::outputDirectory,
        )
    }
}

dependencies {
    implementation(platform(libs.compose.bom))
    implementation(libs.activity.compose)
    implementation(libs.compose.material3)
    implementation(libs.compose.ui)
    implementation(libs.compose.ui.tooling.preview)
    implementation(libs.coroutines.android)
    // Cronet with its native library, so the app doesn't depend on Google Play services or Android 14's
    // HttpEngine
    implementation(libs.cronet.bundled)
    implementation(libs.lifecycle.runtime.compose)
    implementation(libs.lifecycle.viewmodel.compose)
    implementation(libs.protobuf.javalite)
    implementation(libs.trevrpc.core)
    implementation(libs.trevrpc.transport.cronet)
    debugImplementation(libs.compose.ui.tooling)

    testImplementation(libs.coroutines.test)
    testImplementation(libs.junit)
}
