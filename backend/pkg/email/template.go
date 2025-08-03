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

func LoginSuccessTemplate(username, email, loginTime, userAgent, ipAddress string) string {
	return fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<div style="background-color: #f8f9fa; padding: 20px; border-radius: 8px; margin-bottom: 20px;">
					<h2 style="color: #28a745; margin-top: 0;">🔐 Login Successful</h2>
					<p>Hello <strong>%s</strong>,</p>
					<p>We detected a successful login to your account. Here are the details:</p>
				</div>
				
				<div style="background-color: #ffffff; padding: 20px; border: 1px solid #dee2e6; border-radius: 8px; margin-bottom: 20px;">
					<h3 style="color: #495057; margin-top: 0;">📋 Login Details</h3>
					<table style="width: 100%%; border-collapse: collapse;">
						<tr>
							<td style="padding: 8px 0; border-bottom: 1px solid #eee;"><strong>Username:</strong></td>
							<td style="padding: 8px 0; border-bottom: 1px solid #eee;">%s</td>
						</tr>
						<tr>
							<td style="padding: 8px 0; border-bottom: 1px solid #eee;"><strong>Email:</strong></td>
							<td style="padding: 8px 0; border-bottom: 1px solid #eee;">%s</td>
						</tr>
						<tr>
							<td style="padding: 8px 0; border-bottom: 1px solid #eee;"><strong>Login Time:</strong></td>
							<td style="padding: 8px 0; border-bottom: 1px solid #eee;">%s</td>
						</tr>
						<tr>
							<td style="padding: 8px 0; border-bottom: 1px solid #eee;"><strong>IP Address:</strong></td>
							<td style="padding: 8px 0; border-bottom: 1px solid #eee;">%s</td>
						</tr>
						<tr>
							<td style="padding: 8px 0;"><strong>User Agent:</strong></td>
							<td style="padding: 8px 0;">%s</td>
						</tr>
					</table>
				</div>
				
				<div style="background-color: #fff3cd; padding: 15px; border: 1px solid #ffeaa7; border-radius: 8px; margin-bottom: 20px;">
					<h4 style="color: #856404; margin-top: 0;">⚠️ Security Notice</h4>
					<p style="margin-bottom: 0;">If this login was not initiated by you, please:</p>
					<ul style="margin: 10px 0 0 0; padding-left: 20px;">
						<li>Change your password immediately</li>
						<li>Enable two-factor authentication if available</li>
						<li>Contact our support team</li>
					</ul>
				</div>
				
				<div style="text-align: center; padding: 20px; background-color: #f8f9fa; border-radius: 8px;">
					<p style="margin: 0; color: #6c757d; font-size: 14px;">
						This is an automated security notification. Please do not reply to this email.
					</p>
				</div>
			</div>
		</body>
		</html>
	`, username, username, email, loginTime, ipAddress, userAgent)
}

func PasswordResetTemplate(username, resetURL string) string {
	return fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<div style="background-color: #f8f9fa; padding: 20px; border-radius: 8px; margin-bottom: 20px;">
					<h2 style="color: #007bff; margin-top: 0;">🔑 Password Reset Request</h2>
					<p>Hello <strong>%s</strong>,</p>
					<p>We received a request to reset your password. Click the button below to create a new password:</p>
				</div>
				
				<div style="text-align: center; padding: 30px; background-color: #ffffff; border: 1px solid #dee2e6; border-radius: 8px; margin-bottom: 20px;">
					<a href="%s" style="display: inline-block; background-color: #007bff; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; font-weight: bold; font-size: 16px;">
						Reset Password
					</a>
				</div>
				
				<div style="background-color: #fff3cd; padding: 15px; border: 1px solid #ffeaa7; border-radius: 8px; margin-bottom: 20px;">
					<h4 style="color: #856404; margin-top: 0;">⚠️ Security Notice</h4>
					<p style="margin-bottom: 0;">If you didn't request a password reset:</p>
					<ul style="margin: 10px 0 0 0; padding-left: 20px;">
						<li>Ignore this email</li>
						<li>Your password will remain unchanged</li>
						<li>Contact our support team if you have concerns</li>
					</ul>
				</div>
				
				<div style="background-color: #e9ecef; padding: 15px; border-radius: 8px; margin-bottom: 20px;">
					<h4 style="color: #495057; margin-top: 0;">📋 Important Information</h4>
					<ul style="margin: 10px 0 0 0; padding-left: 20px;">
						<li>This link will expire in 1 hour</li>
						<li>You can only use this link once</li>
						<li>Choose a strong, unique password</li>
					</ul>
				</div>
				
				<div style="text-align: center; padding: 20px; background-color: #f8f9fa; border-radius: 8px;">
					<p style="margin: 0; color: #6c757d; font-size: 14px;">
						This is an automated password reset request. Please do not reply to this email.
					</p>
				</div>
			</div>
		</body>
		</html>
	`, username, resetURL)
}
