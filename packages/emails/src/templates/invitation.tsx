import { Body, Container, Head, Heading, Html, Link, Text } from "@react-email/components";

interface InvitationEmailProps {
  workspaceName: string;
  role: string;
  invitationUrl: string;
}

export const InvitationEmail = ({
  workspaceName = "{{.WorkspaceName}}",
  role = "{{.Role}}",
  invitationUrl = "{{.InvitationURL}}",
}: InvitationEmailProps) => (
  <Html>
    <Head />
    <Body style={{ fontFamily: "sans-serif", backgroundColor: "#f5f5f5" }}>
      <Container style={{ backgroundColor: "#ffffff", padding: "32px" }}>
        <Heading>You’re invited to {workspaceName}</Heading>
        <Text>Join this Flux workspace as a {role}.</Text>
        <Link href={invitationUrl}>View invitation</Link>
        <Text>
          This invitation expires in seven days. If you didn’t expect it, you can ignore this email.
        </Text>
      </Container>
    </Body>
  </Html>
);

InvitationEmail.PreviewProps = {
  workspaceName: "Example workspace",
  role: "member",
  invitationUrl: "https://app.flux.test/invitations",
};
export default InvitationEmail;
