# Group Assertions

Check if groups exist or are absent.

## Schema

```yaml
tests:
  groups:
    - name: "Test description"
      groups: [group1, group2]   # required - one or more groups
      state: present|absent       # required (default: present)
```

## Examples

**Group exists:**
```yaml
tests:
  groups:
    - name: "Docker group exists"
      groups:
        - docker
      state: present
```

**Multiple groups exist:**
```yaml
tests:
  groups:
    - name: "Required groups present"
      groups:
        - docker
        - sudo
        - www-data
      state: present
```

**Group absent:**
```yaml
tests:
  groups:
    - name: "Legacy groups removed"
      groups:
        - oldapp
        - deprecated
      state: absent
```
