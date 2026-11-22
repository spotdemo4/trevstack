package zip.trev.template.stack

import java.net.URI
import java.net.URISyntaxException

/**
 * The URL of a server: `https`, a host and an optional port (443 by default), without a path,
 * credentials, a query string or a fragment, as the command-line client accepts.
 */
class ServiceUrl private constructor(
    /** The URL in canonical form, such as `https://localhost:8080`. */
    val origin: String,
) {
    override fun equals(other: Any?): Boolean = other is ServiceUrl && other.origin == origin

    override fun hashCode(): Int = origin.hashCode()

    override fun toString(): String = origin

    companion object {
        private const val DEFAULT_PORT = 443
        private const val MAX_PORT = 65535

        /**
         * Parses a server URL.
         *
         * @throws IllegalArgumentException with a message for the user if [input] is not a valid
         *   server URL.
         */
        fun parse(input: String): ServiceUrl {
            val uri =
                try {
                    URI(input.trim())
                } catch (error: URISyntaxException) {
                    throw IllegalArgumentException("invalid server URL: ${error.reason}", error)
                }
            require(uri.scheme != null) { "server URL must start with https://" }
            require(uri.scheme.equals("https", ignoreCase = true)) { "server URL must use https, not ${uri.scheme}" }
            require(!uri.isOpaque && uri.rawAuthority != null) { "server URL must include a host" }
            require(uri.rawUserInfo == null) { "server URL must not contain credentials" }
            val host = requireNotNull(uri.host?.lowercase()) { "server URL must include a valid host" }
            require(uri.rawPath.isNullOrEmpty() || uri.rawPath == "/") { "server URL must not contain a path" }
            require(uri.rawQuery == null) { "server URL must not contain a query string" }
            require(uri.rawFragment == null) { "server URL must not contain a fragment" }
            val port = if (uri.port == -1) DEFAULT_PORT else uri.port
            require(port in 1..MAX_PORT) { "server URL has an invalid port" }
            return ServiceUrl("https://$host:$port")
        }
    }
}
