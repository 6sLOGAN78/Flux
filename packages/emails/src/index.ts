// Exported transactional email template declarations
export interface WelcomeEmailProps {
  userFirstname: string;
}

export const WelcomeEmailTemplate = ({ userFirstname }: WelcomeEmailProps): string => {
  return `
    <!DOCTYPE html>
    <html>
      <head><title>Welcome to Flux</title></head>
      <body>
        <h1>Welcome to Flux, ${userFirstname}!</h1>
        <p>Your multi-tier link management and enterprise analytics account is ready.</p>
      </body>
    </html>
  `;
};
