package zip.trev.template.stack

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertSame
import org.junit.Test
import zip.trev.template.stack.gen.cursor.v1.Cursor
import zip.trev.template.stack.gen.cursor.v1.ShareResponse
import zip.trev.template.stack.gen.cursor.v1.StreamEnd
import zip.trev.template.stack.gen.cursor.v1.WatchResponse
import kotlin.time.Duration.Companion.seconds
import kotlin.time.TestTimeSource

class CursorsTest {
    private val cursor =
        Cursor
            .newBuilder()
            .setStreamId(7)
            .setUsername("alice")
            .setX(0.25)
            .setY(0.75)
            .build()

    @Test
    fun mapsTouchesToPositions() {
        assertEquals(Position(0.25, 0.5), Position.of(25f, 100f, 100, 200))
        assertEquals(Position(0.0, 1.0), Position.of(-5f, 300f, 100, 200))
        assertEquals(Position(0.0, 0.0), Position.of(10f, 10f, 0, 0))
    }

    @Test
    fun readsEvents() {
        val moved = CursorEvent.Moved(7, "alice", Position(0.25, 0.75))
        assertEquals(moved, CursorEvent.of(WatchResponse.newBuilder().setMoved(cursor).build()))
        assertEquals(moved, CursorEvent.of(ShareResponse.newBuilder().setMoved(cursor).build()))
        val end =
            StreamEnd
                .newBuilder()
                .setStreamId(7)
                .setUsername("alice")
                .build()
        assertEquals(CursorEvent.Ended(7), CursorEvent.of(WatchResponse.newBuilder().setEnded(end).build()))
        assertEquals(CursorEvent.Ended(7), CursorEvent.of(ShareResponse.newBuilder().setEnded(end).build()))
        assertNull(CursorEvent.of(WatchResponse.getDefaultInstance()))
    }

    @Test
    fun movesAndRemovesCursors() {
        val time = TestTimeSource()
        val now = time.markNow()
        val cursors =
            Cursors()
                .apply(CursorEvent.Moved(1, "alice", Position(0.1, 0.1)), now)
                .apply(CursorEvent.Moved(2, "bob", Position(0.2, 0.2)), now)
                .apply(CursorEvent.Moved(1, "alice", Position(0.3, 0.3)), now)
        assertEquals(setOf(1L, 2L), cursors.byStream.keys)
        assertEquals(Position(0.3, 0.3), cursors.byStream.getValue(1).position)
        assertEquals(setOf(2L), cursors.apply(CursorEvent.Ended(1), now).byStream.keys)
    }

    @Test
    fun expiresIdleCursors() {
        val time = TestTimeSource()
        val cursors = Cursors().apply(CursorEvent.Moved(1, "alice", Position(0.1, 0.1)), time.markNow())
        time += 5.seconds
        val moved = cursors.apply(CursorEvent.Moved(2, "bob", Position(0.2, 0.2)), time.markNow())
        assertSame(moved, moved.expire())
        time += 6.seconds
        assertEquals(setOf(2L), moved.expire().byStream.keys)
    }

    @Test
    fun colorsUsernamesAsTheClientDoes() {
        // the command-line client's 64-bit FNV-1a hash of each username, modulo its 6 colours
        val expected = mapOf("" to 5, "alice" to 5, "bob" to 0, "carol" to 4, "demo_user" to 0)
        for ((username, index) in expected) {
            assertEquals(username, index, colorIndex(username, 6))
        }
    }
}
