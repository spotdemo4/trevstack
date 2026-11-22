package zip.trev.template.stack

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.consumeAsFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.job
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import zip.trev.trevrpc.Code
import zip.trev.trevrpc.TrevRpcException
import java.util.Locale
import kotlin.time.Duration.Companion.seconds
import kotlin.time.TimeSource

/** What touching the pad does, one mode for each kind of RPC. */
enum class Mode(
    val label: String,
    val hint: String,
) {
    /** `Send`, unary. */
    SEND("Send", "Tap the pad to record a position"),

    /** `Watch`, server streaming. */
    WATCH("Watch", "Watching every cursor"),

    /** `Track`, client streaming. */
    TRACK("Track", "Drag on the pad to stream your position"),

    /** `Share`, bidirectional streaming. */
    SHARE("Share", "Drag on the pad to share your position"),
}

/** The sign-in form. */
data class SignInForm(
    val url: String = BuildConfig.DEFAULT_URL,
    val username: String = "",
    val busy: Boolean = false,
    val error: String? = null,
)

/** The pad of a signed-in user. */
data class Pad(
    val username: String,
    val url: ServiceUrl,
    val mode: Mode,
    val cursors: Cursors = Cursors(),
    /** Where the user is touching the pad, if they are. */
    val touch: Position? = null,
    val status: String = mode.hint,
)

sealed interface UiState {
    data class SignedOut(
        val form: SignInForm,
    ) : UiState

    data class SignedIn(
        val pad: Pad,
    ) : UiState
}

class StackViewModel(
    application: Application,
) : AndroidViewModel(application) {
    private val mutableState = MutableStateFlow<UiState>(UiState.SignedOut(SignInForm()))
    val state: StateFlow<UiState> = mutableState.asStateFlow()

    private var client: StackClient? = null

    // the calls of the current mode, cancelled when it changes
    private var modeScope: CoroutineScope? = null
    private var trackPositions: Channel<Position>? = null
    private var sharePositions: Channel<Position>? = null

    init {
        viewModelScope.launch {
            while (true) {
                delay(1.seconds)
                updatePad { it.copy(cursors = it.cursors.expire()) }
            }
        }
    }

    fun updateForm(change: (SignInForm) -> SignInForm) {
        mutableState.update { state ->
            if (state is UiState.SignedOut && !state.form.busy) UiState.SignedOut(change(state.form).copy(error = null)) else state
        }
    }

    /** Signs in as the form's user with [password], after creating the account if [createAccount]. */
    fun signIn(
        createAccount: Boolean,
        password: String,
    ) {
        val form = (state.value as? UiState.SignedOut)?.form ?: return
        if (form.busy) return
        val url =
            try {
                ServiceUrl.parse(form.url)
            } catch (error: IllegalArgumentException) {
                return showFormError(error.message ?: "invalid server URL")
            }
        val username = form.username.trim()
        if (username.isEmpty()) return showFormError("Enter a username")
        if (password.isEmpty()) return showFormError("Enter a password")

        mutableState.value = UiState.SignedOut(form.copy(username = username, busy = true, error = null))
        viewModelScope.launch {
            var connecting: StackClient? = null
            try {
                // building the engine loads Cronet's native library
                val connected = withContext(Dispatchers.IO) { StackClient(getApplication(), url) }
                connecting = connected
                if (createAccount) connected.signup(username, password)
                val subject = connected.login(username, password)
                client = connected
                mutableState.value = UiState.SignedIn(Pad(subject, url, Mode.WATCH))
                setMode(Mode.WATCH)
            } catch (error: CancellationException) {
                connecting?.let { release(it, logout = false) }
                throw error
            } catch (error: Exception) {
                connecting?.let { release(it, logout = false) }
                val action = if (createAccount) "Sign up" else "Sign in"
                mutableState.value =
                    UiState.SignedOut(form.copy(username = username, busy = false, error = "$action failed: ${describe(error)}"))
            }
        }
    }

    /** Ends the session and returns to the sign-in form, showing [error] if there is one. */
    fun signOut(error: String? = null) {
        val pad = (state.value as? UiState.SignedIn)?.pad ?: return
        stopMode()
        client?.let { release(it, logout = true) }
        client = null
        mutableState.value = UiState.SignedOut(SignInForm(url = pad.url.origin, username = pad.username, error = error))
    }

    /** Switches to [mode], ending the calls of the last one. */
    fun setMode(mode: Mode) {
        val client = client ?: return
        stopMode()
        val scope = CoroutineScope(viewModelScope.coroutineContext + SupervisorJob(viewModelScope.coroutineContext.job))
        modeScope = scope
        updatePad { Pad(it.username, it.url, mode) }
        when (mode) {
            Mode.SEND, Mode.TRACK -> {
                Unit
            }

            Mode.WATCH -> {
                scope.launch {
                    call("Watch") {
                        client.watch().collect(::show)
                        status("The server ended the Watch stream")
                    }
                }
            }

            Mode.SHARE -> {
                val positions = Channel<Position>(Channel.UNLIMITED)
                sharePositions = positions
                scope.launch {
                    call("Share") {
                        client.share(positions.consumeAsFlow()).collect(::show)
                        status("The server ended the Share stream")
                    }
                }
            }
        }
    }

    /** The user touched the pad at [position]. */
    fun press(position: Position) {
        val pad = (state.value as? UiState.SignedIn)?.pad ?: return
        val client = client ?: return
        val scope = modeScope ?: return
        when (pad.mode) {
            Mode.SEND -> {
                updatePad { it.copy(touch = position, status = "Sending ${format(position)}…") }
                scope.launch {
                    call("Send") {
                        val streamId = client.send(position)
                        status("Recorded ${format(position)} as stream $streamId")
                    }
                }
            }

            Mode.WATCH -> {
                Unit
            }

            // each drag is a Track stream of its own
            Mode.TRACK -> {
                trackPositions?.close()
                val positions = Channel<Position>(Channel.UNLIMITED)
                trackPositions = positions
                positions.trySend(position)
                updatePad { it.copy(touch = position, status = "Tracking…") }
                scope.launch {
                    call("Track") {
                        val summary = client.track(positions.consumeAsFlow())
                        status(
                            if (summary == null) {
                                "Tracked no positions"
                            } else {
                                String.format(
                                    Locale.ROOT,
                                    "Tracked %d positions over %.4f as stream %d",
                                    summary.count,
                                    summary.distance,
                                    summary.id,
                                )
                            },
                        )
                    }
                }
            }

            Mode.SHARE -> {
                move(position)
            }
        }
    }

    /** The user moved their finger on the pad to [position]. */
    fun move(position: Position) {
        val pad = (state.value as? UiState.SignedIn)?.pad ?: return
        val positions =
            when (pad.mode) {
                Mode.TRACK -> trackPositions
                Mode.SHARE -> sharePositions
                Mode.SEND, Mode.WATCH -> null
            } ?: return
        positions.trySend(position)
        updatePad { it.copy(touch = position) }
    }

    /** The user lifted their finger off the pad. */
    fun release() {
        // ends the Track stream, whose summary the server then returns
        trackPositions?.close()
        trackPositions = null
        updatePad { if (it.mode == Mode.SEND) it else it.copy(touch = null) }
    }

    override fun onCleared() {
        stopMode()
        client?.let { release(it, logout = true) }
        client = null
    }

    private fun stopMode() {
        modeScope?.cancel()
        modeScope = null
        trackPositions?.close()
        trackPositions = null
        sharePositions?.close()
        sharePositions = null
    }

    /** Runs the RPC [name], showing how it failed if it does. */
    private suspend fun call(
        name: String,
        block: suspend () -> Unit,
    ) {
        try {
            block()
        } catch (error: CancellationException) {
            throw error
        } catch (error: Exception) {
            if (error is TrevRpcException && error.status.code == Code.UNAUTHENTICATED) {
                signOut("Your session ended, sign in again")
            } else {
                status("$name failed: ${describe(error)}")
            }
        }
    }

    private fun show(event: CursorEvent) {
        val now = TimeSource.Monotonic.markNow()
        updatePad { it.copy(cursors = it.cursors.apply(event, now)) }
    }

    private fun status(status: String) = updatePad { it.copy(status = status) }

    private fun showFormError(error: String) {
        mutableState.update { state -> if (state is UiState.SignedOut) UiState.SignedOut(state.form.copy(error = error)) else state }
    }

    private fun updatePad(change: (Pad) -> Pad) {
        mutableState.update { state -> if (state is UiState.SignedIn) UiState.SignedIn(change(state.pad)) else state }
    }

    companion object {
        /**
         * Logs out and closes clients after the view model that used them is gone, so the session
         * ends even when the app is closed.
         */
        private val cleanup = CoroutineScope(SupervisorJob() + Dispatchers.IO)

        private fun release(
            client: StackClient,
            logout: Boolean,
        ) {
            cleanup.launch {
                if (logout) runCatching { client.logout() }
                runCatching { client.close() }
            }
        }
    }
}

/** Describes why a call failed, for the user. */
fun describe(error: Throwable): String {
    if (error !is TrevRpcException) return error.message ?: error.javaClass.simpleName
    val status = error.status
    val description = status.code.name.lowercase(Locale.ROOT)
    val message = if (status.message.isEmpty()) description else "$description: ${status.message}"
    val retry = error.retryDelay ?: return message
    return "$message, try again in ${retry.inWholeSeconds.coerceAtLeast(1)}s"
}

/** Formats a position as the command-line client does. */
fun format(position: Position): String = String.format(Locale.ROOT, "(%.4f, %.4f)", position.x, position.y)
