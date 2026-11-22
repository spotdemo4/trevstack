package zip.trev.template.stack

import zip.trev.template.stack.gen.cursor.v1.Cursor
import zip.trev.template.stack.gen.cursor.v1.ShareResponse
import zip.trev.template.stack.gen.cursor.v1.StreamEnd
import zip.trev.template.stack.gen.cursor.v1.WatchResponse
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds
import kotlin.time.TimeMark

/** How long a cursor stays on the pad without moving, as in the command-line client. */
val IDLE_TIMEOUT: Duration = 10.seconds

/**
 * A position on the pad, as fractions of its width and height from its top-left corner, so every
 * client shows it in the same place whatever its size.
 */
data class Position(
    val x: Double,
    val y: Double,
) {
    companion object {
        /** The position of a point [x], [y] on a pad [width] by [height], clamped to the pad. */
        fun of(
            x: Float,
            y: Float,
            width: Int,
            height: Int,
        ): Position = Position(fraction(x, width), fraction(y, height))

        private fun fraction(
            offset: Float,
            size: Int,
        ): Double = if (size <= 0) 0.0 else (offset.toDouble() / size).coerceIn(0.0, 1.0)
    }
}

/** A cursor event from `Watch` or `Share`. */
sealed interface CursorEvent {
    val streamId: Long

    /** A stream's cursor moved. */
    data class Moved(
        override val streamId: Long,
        val username: String,
        val position: Position,
    ) : CursorEvent

    /** A stream ended, so its cursor can be removed. */
    data class Ended(
        override val streamId: Long,
    ) : CursorEvent

    companion object {
        /** The event of a `Watch` response, or null for one this app doesn't know. */
        fun of(response: WatchResponse): CursorEvent? =
            when (response.eventCase) {
                WatchResponse.EventCase.MOVED -> moved(response.moved)
                WatchResponse.EventCase.ENDED -> ended(response.ended)
                else -> null
            }

        /** The event of a `Share` response, or null for one this app doesn't know. */
        fun of(response: ShareResponse): CursorEvent? =
            when (response.eventCase) {
                ShareResponse.EventCase.MOVED -> moved(response.moved)
                ShareResponse.EventCase.ENDED -> ended(response.ended)
                else -> null
            }

        private fun moved(cursor: Cursor) = Moved(cursor.streamId, cursor.username, Position(cursor.x, cursor.y))

        private fun ended(end: StreamEnd) = Ended(end.streamId)
    }
}

/** Another stream's cursor on the pad. */
data class RemoteCursor(
    val username: String,
    val position: Position,
    val moved: TimeMark,
)

/** The cursors on the pad, by stream ID. */
data class Cursors(
    val byStream: Map<Long, RemoteCursor> = emptyMap(),
) {
    /** Moves the cursor of a [CursorEvent.Moved] event's stream to [now], or removes an ended one. */
    fun apply(
        event: CursorEvent,
        now: TimeMark,
    ): Cursors =
        when (event) {
            is CursorEvent.Moved -> Cursors(byStream + (event.streamId to RemoteCursor(event.username, event.position, now)))
            is CursorEvent.Ended -> Cursors(byStream - event.streamId)
        }

    /** Removes the cursors that have not moved for [IDLE_TIMEOUT]. */
    fun expire(): Cursors {
        val active = byStream.filterValues { it.moved.elapsedNow() < IDLE_TIMEOUT }
        return if (active.size == byStream.size) this else Cursors(active)
    }
}

/**
 * The index into a palette of [size] colours of a username's cursor, the same on every run and in
 * the command-line client.
 */
fun colorIndex(
    username: String,
    size: Int,
): Int {
    // 64-bit FNV-1a, the command-line client's hash
    var hash = -0x340d631b7bdddcdbL // 0xcbf29ce484222325
    for (byte in username.encodeToByteArray()) {
        hash = (hash xor (byte.toLong() and 0xff)) * 0x100000001b3L
    }
    return hash.toULong().rem(size.toULong()).toInt()
}
