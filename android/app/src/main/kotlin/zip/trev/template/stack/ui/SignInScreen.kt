package zip.trev.template.stack.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.text.input.rememberTextFieldState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedSecureTextField
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import zip.trev.template.stack.SignInForm

@Composable
fun SignInScreen(
    form: SignInForm,
    onChange: ((SignInForm) -> SignInForm) -> Unit,
    onSignIn: (createAccount: Boolean, password: String) -> Unit,
) {
    // kept out of the view model, which outlives the screen
    val password = rememberTextFieldState()
    val signIn = { createAccount: Boolean -> onSignIn(createAccount, password.text.toString()) }
    Surface(modifier = Modifier.fillMaxSize()) {
        Column(
            modifier =
                Modifier
                    .safeDrawingPadding()
                    .imePadding()
                    .verticalScroll(rememberScrollState())
                    .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            Column(
                modifier = Modifier.widthIn(max = 480.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp),
            ) {
                Text("Sign in", style = MaterialTheme.typography.headlineMedium)
                OutlinedTextField(
                    value = form.url,
                    onValueChange = { url -> onChange { it.copy(url = url) } },
                    label = { Text("Server") },
                    singleLine = true,
                    enabled = !form.busy,
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Uri, imeAction = ImeAction.Next),
                    modifier = Modifier.fillMaxWidth(),
                )
                OutlinedTextField(
                    value = form.username,
                    onValueChange = { username -> onChange { it.copy(username = username) } },
                    label = { Text("Username") },
                    singleLine = true,
                    enabled = !form.busy,
                    keyboardOptions = KeyboardOptions(autoCorrectEnabled = false, imeAction = ImeAction.Next),
                    modifier = Modifier.fillMaxWidth(),
                )
                OutlinedSecureTextField(
                    state = password,
                    label = { Text("Password") },
                    enabled = !form.busy,
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password, imeAction = ImeAction.Go),
                    onKeyboardAction = { signIn(false) },
                    modifier = Modifier.fillMaxWidth(),
                )
                form.error?.let { error ->
                    Text(error, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodyMedium)
                }
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp, Alignment.End),
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    if (form.busy) CircularProgressIndicator(modifier = Modifier.padding(end = 8.dp))
                    OutlinedButton(onClick = { signIn(true) }, enabled = !form.busy) { Text("Sign up") }
                    Button(onClick = { signIn(false) }, enabled = !form.busy) { Text("Sign in") }
                }
            }
        }
    }
}

@Preview
@Composable
private fun SignInScreenPreview() {
    StackTheme(dynamicColor = false) {
        SignInScreen(SignInForm(username = "demo_user", error = "Sign in failed: unauthenticated"), {}, { _, _ -> })
    }
}
