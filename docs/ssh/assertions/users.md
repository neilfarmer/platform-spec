# User Assertions

Check user properties including shell, home directory, and group membership.

## Schema

```yaml
tests:
  users:
    - name: "Test description"
      user: "username"           # required
      shell: "/bin/bash"         # optional
      home: "/home/user"         # optional
      groups: [group1, group2]   # optional
```

## Examples

**User exists:**
```yaml
tests:
  users:
    - name: "Ubuntu user exists"
      user: ubuntu
```

**User with shell:**
```yaml
tests:
  users:
    - name: "Root uses bash"
      user: root
      shell: /bin/bash
```

**User with home directory:**
```yaml
tests:
  users:
    - name: "App user home"
      user: appuser
      home: /opt/app
```

**User with group membership:**
```yaml
tests:
  users:
    - name: "Deploy user in correct groups"
      user: deploy
      groups:
        - docker
        - sudo
```

**Complete example:**
```yaml
tests:
  users:
    - name: "Application user configured"
      user: appuser
      shell: /bin/bash
      home: /opt/app
      groups:
        - docker
        - www-data
```
