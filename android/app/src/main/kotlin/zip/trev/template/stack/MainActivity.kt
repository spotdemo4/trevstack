package zip.trev.template.stack

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.viewModels
import androidx.compose.runtime.getValue
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import zip.trev.template.stack.ui.PadScreen
import zip.trev.template.stack.ui.SignInScreen
import zip.trev.template.stack.ui.StackTheme

class MainActivity : ComponentActivity() {
    private val viewModel: StackViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            StackTheme {
                val state by viewModel.state.collectAsStateWithLifecycle()
                when (val current = state) {
                    is UiState.SignedOut -> {
                        SignInScreen(
                            form = current.form,
                            onChange = viewModel::updateForm,
                            onSignIn = viewModel::signIn,
                        )
                    }

                    is UiState.SignedIn -> {
                        PadScreen(
                            pad = current.pad,
                            onMode = viewModel::setMode,
                            onPress = viewModel::press,
                            onMove = viewModel::move,
                            onRelease = viewModel::release,
                            onSignOut = { viewModel.signOut() },
                        )
                    }
                }
            }
        }
    }
}
