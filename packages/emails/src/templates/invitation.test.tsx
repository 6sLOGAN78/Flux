import assert from "node:assert/strict";
import { test } from "node:test";
import { render } from "@react-email/components";
import InvitationEmail from "./invitation.js";

test("invitation escapes names, roles and href attributes without changing the fixed origin", async () => {
  const html = await render(
    <InvitationEmail
      workspaceName={"<script>private-name</script>&"}
      role={'member"<img>'}
      invitationUrl={'https://app.flux.test/invitations#token=opaque&x="<'}
    />,
  );
  assert.doesNotMatch(html, /<script>|<img>/);
  assert.match(html, /&lt;script&gt;private-name&lt;\/script&gt;&amp;/);
  assert.match(html, /member&quot;&lt;img&gt;/);
  assert.match(html, /href="https:\/\/app\.flux\.test\/invitations#token=opaque&amp;x=&quot;&lt;"/);
});

test("invitation export preserves all Go template fields deterministically", async () => {
  const props = {
    workspaceName: "{{.WorkspaceName}}",
    role: "{{.Role}}",
    invitationUrl: "{{.InvitationURL}}",
  };
  const first = await render(<InvitationEmail {...props} />, { pretty: true });
  assert.equal(first, await render(<InvitationEmail {...props} />, { pretty: true }));
  for (const field of Object.values(props)) assert.ok(first.includes(field));
});
