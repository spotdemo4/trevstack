package zip.trev.template.stack

import org.junit.Assert.assertEquals
import org.junit.Assert.assertThrows
import org.junit.Test

class ServiceUrlTest {
    @Test
    fun parsesOrigins() {
        assertEquals("https://localhost:8080", ServiceUrl.parse("https://localhost:8080").origin)
        assertEquals("https://example.com:443", ServiceUrl.parse(" HTTPS://Example.com/ ").origin)
        assertEquals("https://127.0.0.1:8443", ServiceUrl.parse("https://127.0.0.1:8443").origin)
        assertEquals("https://[::1]:8080", ServiceUrl.parse("https://[::1]:8080").origin)
    }

    @Test
    fun rejectsOtherUrls() {
        for (url in listOf(
            "",
            "localhost:8080",
            "http://localhost:8080",
            "https://",
            "https://user:password@localhost",
            "https://localhost/api",
            "https://localhost?query",
            "https://localhost#fragment",
            "https://localhost:70000",
            "https://local host",
        )) {
            assertThrows(url, IllegalArgumentException::class.java) { ServiceUrl.parse(url) }
        }
    }
}
