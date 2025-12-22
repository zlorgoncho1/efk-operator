# Security Policy

## Supported Versions

We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability, please follow these steps:

### 1. Do NOT create a public GitHub issue

Security vulnerabilities should be reported privately to protect users.

### 2. Report the vulnerability

**Do NOT create a public GitHub issue for security vulnerabilities.**

Instead, please create a **private security advisory** on GitHub:

1. Go to the [Security tab](https://github.com/zlorgoncho1/efk-operator/security) in the repository
2. Click on "Advisories"
3. Click "New draft security advisory"
4. Fill in the details:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if you have one)
5. Submit the draft advisory

**Alternative**: If you prefer, you can create a **private Pull Request** with the prefix `[SECURITY]` in the title. The PR will be reviewed privately before being merged or made public.

Include the following information:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if you have one)

### 3. Response timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity, but we aim to address critical issues as quickly as possible

### 4. Disclosure

We will:
- Acknowledge receipt of your report
- Keep you informed of our progress
- Credit you in the security advisory (if you wish)
- Notify you when the fix is released

## Security Best Practices

When using this operator:

1. **Keep the operator updated**: Always use the latest stable version
2. **Review RBAC permissions**: Ensure the operator has only necessary permissions
3. **Secure your cluster**: Follow Kubernetes security best practices
4. **Use TLS**: Enable TLS for all components in production
5. **Enable authentication**: Use authentication for Elasticsearch and Kibana
6. **Network policies**: Implement network policies to restrict traffic
7. **Secrets management**: Use Kubernetes secrets or external secret management systems
8. **Regular audits**: Regularly audit your deployments and configurations

## Known Security Considerations

### Operator Permissions

The operator requires certain RBAC permissions to function:
- Read/write access to EFKStack custom resources
- Ability to create/update/delete Helm releases
- Access to create Kubernetes resources (StatefulSets, Deployments, Services, etc.)

Review the RBAC configuration in `config/rbac/` and adjust as needed for your security requirements.

### Helm Charts

The Helm charts deploy components with default configurations. For production:
- Review and customize security settings
- Enable TLS and authentication
- Configure network policies
- Use appropriate resource limits
- Review and restrict service accounts

### Container Images

We use official images from:
- Elasticsearch: `docker.elastic.co/elasticsearch/elasticsearch`
- Fluent Bit: `fluent/fluent-bit`
- Kibana: `docker.elastic.co/kibana/kibana`

Always verify image integrity and consider using image scanning tools.

## Security Updates

Security updates will be:
- Released as patch versions (e.g., 0.1.1, 0.1.2)
- Documented in CHANGELOG.md
- Tagged with security labels in GitHub releases
- Announced in release notes

## Additional Resources

- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/)
- [CNCF Security Best Practices](https://www.cncf.io/blog/2021/08/04/kubernetes-security-best-practices/)
- [OWASP Kubernetes Top 10](https://owasp.org/www-project-kubernetes-top-ten/)

## Contact

For security-related questions or concerns, please use the [GitHub Security Advisories](https://github.com/zlorgoncho1/efk-operator/security/advisories) or create a private Pull Request with the `[SECURITY]` prefix.

