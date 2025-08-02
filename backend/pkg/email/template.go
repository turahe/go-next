package email

import "fmt"

func UserWelcomeTemplate(username string) string {
	return fmt.Sprintf(`
		<html>
		<body>
			<h2>Welcome, %s!</h2>
			<p>Thank you for registering. We're glad to have you on board.</p>
		</body>
		</html>
	`, username)
}

func EmailVerificationTemplate(username, verifyURL string) string {
	return fmt.Sprintf(`
		<html>
		<body>
			<h2>Hello, %s!</h2>
			<p>Thank you for registering. Please verify your email address by clicking the link below:</p>
			<p><a href="%s">Verify Email</a></p>
			<p>If you did not register, please ignore this email.</p>
		</body>
		</html>
	`, username, verifyURL)
}

func PhoneVerificationTemplate(username, code string) string {
	return fmt.Sprintf(`
		Hello, %s!\nYour phone verification code is: %s\nIf you did not request this, please ignore this message.
	`, username, code)
}
