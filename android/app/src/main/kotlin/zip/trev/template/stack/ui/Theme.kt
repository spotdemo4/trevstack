package zip.trev.template.stack.ui

import android.os.Build
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.dynamicDarkColorScheme
import androidx.compose.material3.dynamicLightColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext

// Catppuccin, as in the web app: Mocha when dark and Latte when light
private val dark =
    darkColorScheme(
        primary = Color(0xFFB4BEFE),
        onPrimary = Color(0xFF1E1E2E),
        background = Color(0xFF1E1E2E),
        onBackground = Color(0xFFCDD6F4),
        surface = Color(0xFF1E1E2E),
        onSurface = Color(0xFFCDD6F4),
        surfaceVariant = Color(0xFF313244),
        onSurfaceVariant = Color(0xFFBAC2DE),
        error = Color(0xFFF38BA8),
    )

private val light =
    lightColorScheme(
        primary = Color(0xFF7287FD),
        onPrimary = Color(0xFFEFF1F5),
        background = Color(0xFFEFF1F5),
        onBackground = Color(0xFF4C4F69),
        surface = Color(0xFFEFF1F5),
        onSurface = Color(0xFF4C4F69),
        surfaceVariant = Color(0xFFCCD0DA),
        onSurfaceVariant = Color(0xFF5C5F77),
        error = Color(0xFFD20F39),
    )

/** The colours of other users' cursors, in the command-line client's order: Mocha when dark. */
private val darkCursorColors =
    listOf(
        Color(0xFFF38BA8), // red
        Color(0xFFA6E3A1), // green
        Color(0xFFF9E2AF), // yellow
        Color(0xFF89B4FA), // blue
        Color(0xFFCBA6F7), // mauve
        Color(0xFF94E2D5), // teal
    )

/** Latte when light, which stand out on its paler background. */
private val lightCursorColors =
    listOf(
        Color(0xFFD20F39), // red
        Color(0xFF40A02B), // green
        Color(0xFFDF8E1D), // yellow
        Color(0xFF1E66F5), // blue
        Color(0xFF8839EF), // mauve
        Color(0xFF179299), // teal
    )

/** The colours of other users' cursors in the current theme. */
@Composable
fun cursorColors(): List<Color> = if (isSystemInDarkTheme()) darkCursorColors else lightCursorColors

@Composable
fun StackTheme(
    dynamicColor: Boolean = true,
    content: @Composable () -> Unit,
) {
    val darkTheme = isSystemInDarkTheme()
    val colorScheme =
        when {
            dynamicColor && Build.VERSION.SDK_INT >= Build.VERSION_CODES.S -> {
                val context = LocalContext.current
                if (darkTheme) dynamicDarkColorScheme(context) else dynamicLightColorScheme(context)
            }

            darkTheme -> {
                dark
            }

            else -> {
                light
            }
        }
    MaterialTheme(colorScheme = colorScheme, content = content)
}
