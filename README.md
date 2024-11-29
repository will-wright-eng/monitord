# monitord

## setup

`~/.config/monitord/config.json`

```json
{
  "database": {
    "path": ".config/monitord/monitord.db"
  },
  "monitor": {
    "config_check_interval": "180s",
    "endpoints": [
      {
        "name": "Cyber Epistemics",
        "url": "https://cyberepistemics.com",
        "interval": "60s",
        "timeout": "10s",
        "description": "Cyber Epistemics Personal Website",
        "tags": ["production", "external", "blog"],
        "enabled": true
      }
    ]
  },
  "logging": {
    "path": "/usr/local/var/log/monitord.log",
    "level": "info"
  }
}
```
