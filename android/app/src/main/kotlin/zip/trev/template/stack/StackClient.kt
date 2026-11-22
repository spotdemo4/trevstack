package zip.trev.template.stack

import android.content.Context
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.mapNotNull
import org.chromium.net.CronetEngine
import zip.trev.template.stack.gen.auth.v1.AuthServiceClient
import zip.trev.template.stack.gen.auth.v1.LoginRequest
import zip.trev.template.stack.gen.auth.v1.LogoutRequest
import zip.trev.template.stack.gen.auth.v1.SignupRequest
import zip.trev.template.stack.gen.cursor.v1.CursorServiceClient
import zip.trev.template.stack.gen.cursor.v1.SendRequest
import zip.trev.template.stack.gen.cursor.v1.ShareRequest
import zip.trev.template.stack.gen.cursor.v1.StreamSummary
import zip.trev.template.stack.gen.cursor.v1.TrackRequest
import zip.trev.template.stack.gen.cursor.v1.WatchRequest
import zip.trev.trevrpc.CallOptions
import zip.trev.trevrpc.Metadata
import zip.trev.trevrpc.RpcChannel
import zip.trev.trevrpc.cronet.CronetRpcChannel
import java.util.concurrent.ExecutorService
import java.util.concurrent.Executors
import kotlin.time.Duration.Companion.seconds

/** The timeout of each unary call, as in the command-line client. */
private val RPC_TIMEOUT = 10.seconds

/**
 * A connection to a server, over HTTP/3 or, when UDP is blocked, HTTP/2. Calls made after [login]
 * carry its session token as `authorization: Bearer <token>` metadata.
 */
class StackClient(
    context: Context,
    val url: ServiceUrl,
) {
    // the client owns the engine and executor, which the channel borrows
    private val engine: CronetEngine =
        CronetEngine
            .Builder(context.applicationContext)
            .enableQuic(true)
            .enableHttp2(true)
            .build()
    private val executor: ExecutorService = Executors.newSingleThreadExecutor()
    private val channel: RpcChannel = CronetRpcChannel.create(engine, url.origin, executor)
    private val unary = CallOptions(timeout = RPC_TIMEOUT)

    private val auth = AuthServiceClient(channel, unary)

    // streams run for as long as they are used, so only unary calls have a timeout
    private val cursor = CursorServiceClient(channel)

    /** Creates an account. It doesn't sign in. */
    suspend fun signup(
        username: String,
        password: String,
    ) {
        auth.signup(
            SignupRequest
                .newBuilder()
                .setUsername(username)
                .setPassword(password)
                .build(),
        )
    }

    /** Signs in, and returns the session's canonical username. */
    suspend fun login(
        username: String,
        password: String,
    ): String {
        val response =
            auth.login(
                LoginRequest
                    .newBuilder()
                    .setUsername(username)
                    .setPassword(password)
                    .build(),
            )
        channel.setConnectionMetadata(Metadata.bearer(response.token))
        return response.sub
    }

    /** Ends the session. Later calls carry no token. */
    suspend fun logout() {
        try {
            auth.logout(LogoutRequest.getDefaultInstance())
        } finally {
            channel.setConnectionMetadata(Metadata.EMPTY)
        }
    }

    /** Records one position, and returns its stream ID. */
    suspend fun send(position: Position): Long =
        cursor
            .send(
                SendRequest
                    .newBuilder()
                    .setX(position.x)
                    .setY(position.y)
                    .build(),
            ) { unary }
            .cursor
            .streamId

    /** Streams every user's cursor events as they happen. */
    fun watch(): Flow<CursorEvent> = cursor.watch(WatchRequest.getDefaultInstance()).mapNotNull(CursorEvent::of)

    /**
     * Records each of [positions] as one stream, and returns its summary once [positions] ends, or
     * null when it was empty.
     */
    suspend fun track(positions: Flow<Position>): StreamSummary? {
        val response =
            cursor.track(
                positions.map {
                    TrackRequest
                        .newBuilder()
                        .setX(it.x)
                        .setY(it.y)
                        .build()
                },
            )
        return if (response.hasStream()) response.stream else null
    }

    /**
     * Records each of [positions] as one stream, while streaming every other stream's cursor
     * events. The call ends once [positions] ends and the server has sent its last event.
     */
    fun share(positions: Flow<Position>): Flow<CursorEvent> =
        cursor
            .share(
                positions.map {
                    ShareRequest
                        .newBuilder()
                        .setX(it.x)
                        .setY(it.y)
                        .build()
                },
            ).mapNotNull(CursorEvent::of)

    /** Waits for every call to settle, then releases the engine and its executor. */
    suspend fun close() {
        channel.close()
        engine.shutdown()
        executor.shutdown()
    }
}
