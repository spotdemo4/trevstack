package zip.trev.template.stack.ui

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.awaitEachGesture
import androidx.compose.foundation.gestures.awaitFirstDown
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SegmentedButton
import androidx.compose.material3.SegmentedButtonDefaults
import androidx.compose.material3.SingleChoiceSegmentedButtonRow
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clipToBounds
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.drawscope.DrawScope
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.input.pointer.positionChanged
import androidx.compose.ui.text.TextMeasurer
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.drawText
import androidx.compose.ui.text.rememberTextMeasurer
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import zip.trev.template.stack.CursorEvent
import zip.trev.template.stack.Cursors
import zip.trev.template.stack.Mode
import zip.trev.template.stack.Pad
import zip.trev.template.stack.Position
import zip.trev.template.stack.ServiceUrl
import zip.trev.template.stack.colorIndex
import kotlin.time.TimeSource

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PadScreen(
    pad: Pad,
    onMode: (Mode) -> Unit,
    onPress: (Position) -> Unit,
    onMove: (Position) -> Unit,
    onRelease: () -> Unit,
    onSignOut: () -> Unit,
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(pad.username) },
                actions = { TextButton(onClick = onSignOut) { Text("Sign out") } },
            )
        },
    ) { padding ->
        Column(
            modifier =
                Modifier
                    .padding(padding)
                    .padding(horizontal = 16.dp, vertical = 8.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            SingleChoiceSegmentedButtonRow(modifier = Modifier.fillMaxWidth()) {
                Mode.entries.forEachIndexed { index, mode ->
                    SegmentedButton(
                        selected = pad.mode == mode,
                        onClick = { onMode(mode) },
                        shape = SegmentedButtonDefaults.itemShape(index, Mode.entries.size),
                        label = { Text(mode.label) },
                    )
                }
            }
            PadCanvas(
                cursors = pad.cursors,
                touch = pad.touch,
                onPress = onPress,
                onMove = onMove,
                onRelease = onRelease,
                modifier = Modifier.weight(1f).fillMaxWidth(),
            )
            Text(
                pad.status,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}

/** The pad: other streams' cursors, and where the user is touching it. */
@Composable
private fun PadCanvas(
    cursors: Cursors,
    touch: Position?,
    onPress: (Position) -> Unit,
    onMove: (Position) -> Unit,
    onRelease: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val press by rememberUpdatedState(onPress)
    val move by rememberUpdatedState(onMove)
    val release by rememberUpdatedState(onRelease)
    val textMeasurer = rememberTextMeasurer()
    val label = MaterialTheme.typography.labelMedium
    val own = MaterialTheme.colorScheme.primary
    val colors = cursorColors()

    Canvas(
        modifier =
            modifier
                .clipToBounds()
                .background(MaterialTheme.colorScheme.surfaceVariant, RoundedCornerShape(12.dp))
                .pointerInput(Unit) {
                    awaitEachGesture {
                        val down = awaitFirstDown()
                        down.consume()
                        press(Position.of(down.position.x, down.position.y, size.width, size.height))
                        while (true) {
                            val change = awaitPointerEvent().changes.firstOrNull { it.id == down.id } ?: break
                            if (!change.pressed) break
                            if (change.positionChanged()) {
                                change.consume()
                                move(Position.of(change.position.x, change.position.y, size.width, size.height))
                            }
                        }
                        release()
                    }
                },
    ) {
        for (cursor in cursors.byStream.values) {
            val color = colors[colorIndex(cursor.username, colors.size)]
            marker(cursor.position, cursor.username, color, filled = true, textMeasurer, label)
        }
        touch?.let { marker(it, "you", own, filled = false, textMeasurer, label) }
    }
}

private fun DrawScope.marker(
    position: Position,
    name: String,
    color: Color,
    filled: Boolean,
    textMeasurer: TextMeasurer,
    style: TextStyle,
) {
    val radius = 8.dp.toPx()
    val center = Offset((position.x * size.width).toFloat(), (position.y * size.height).toFloat())
    if (filled) {
        drawCircle(color, radius, center)
    } else {
        drawCircle(color, radius, center, style = Stroke(width = 2.dp.toPx()))
    }
    val text = textMeasurer.measure(name, style.copy(color = color))
    // beside the marker, and inside the pad near its right and bottom edges
    val x = (center.x + radius * 1.5f).coerceAtMost(size.width - text.size.width).coerceAtLeast(0f)
    val y = (center.y - text.size.height / 2f).coerceIn(0f, (size.height - text.size.height).coerceAtLeast(0f))
    drawText(text, topLeft = Offset(x, y))
}

@Preview
@Composable
private fun PadScreenPreview() {
    val now = TimeSource.Monotonic.markNow()
    val cursors =
        Cursors()
            .apply(CursorEvent.Moved(1, "alice", Position(0.25, 0.3)), now)
            .apply(CursorEvent.Moved(2, "bob", Position(0.7, 0.6)), now)
    StackTheme(dynamicColor = false) {
        PadScreen(
            pad = Pad("demo_user", ServiceUrl.parse("https://localhost:8080"), Mode.SHARE, cursors, Position(0.5, 0.8)),
            onMode = {},
            onPress = {},
            onMove = {},
            onRelease = {},
            onSignOut = {},
        )
    }
}
