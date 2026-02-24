# Security Policy

## Supported Versions

Security fixes are provided for the actively maintained MVP line.

| Version | Supported |
| --- | --- |
| `main` (latest) | Yes |
| older feature branches | No |
| unmaintained forks | No |

## Reporting a Vulnerability

Please report vulnerabilities privately.

- Github: `@AlhikamWarsawa`
- Subject: `[Kids Planet Security] <short title>`
- Include:
  - affected component(s) (`/api`, `/games`, upload, auth, infra)
  - reproduction steps / proof of concept
  - impact assessment
  - any logs and `request_id` values

Do not disclose vulnerability details publicly until coordinated disclosure is complete.

## Disclosure Policy

Process:
1. We acknowledge receipt.
2. We triage severity and scope.
3. We validate and reproduce.
4. We develop and test a fix.
5. We coordinate release and disclosure timing.

We may request additional technical details to reproduce safely.

## Scope of Security Coverage

In-scope areas include:
- API endpoints and auth flows (`admin`, `player`, `play_token`)
- ZIP upload pipeline and extraction safety
- leaderboard anti-cheat/rate-limit logic
- analytics ingest validation
- MinIO games bucket exposure through `/games`
- Nginx proxy/security headers/CORS behavior
- request correlation and error handling (`request_id`)

Out-of-scope by default:
- vulnerabilities in third-party services outside project control
- social engineering/phishing unrelated to repository code
- best-practice recommendations without demonstrable impact

## Response Timeline (Targets)

- Initial acknowledgment: within 2 business days
- Triage decision: within 5 business days
- Status update cadence: at least weekly for active reports
- Critical fix target: as fast as possible, typically within 7 business days after confirmed reproduction

Timelines are targets, not guarantees, and may vary with complexity.

## Safe Harbor

If you act in good faith and follow this policy, we will not pursue legal action
for security research performed under these guidelines:
- avoid privacy violations, data destruction, or service disruption
- test only against systems you are authorized to assess
- do not exfiltrate more data than needed to demonstrate impact
- provide us a reasonable opportunity to remediate before public disclosure

We appreciate responsible disclosure and will credit researchers when requested
and appropriate.
