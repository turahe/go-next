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
