# GOCursorGen

An independent Go CLI (and soon to be Library) for creating or managing X11 Cursors set

Implementation based on : [xcursorgen](https://gitlab.freedesktop.org/xorg/app/xcursorgen/) ideas

Support file Type: `.png` `.gif` `.jpg` `.ani`

## YAML Schema

```yaml
# Optional Resize
global:
  size: [32, 48, 64, 96]

cursor:
    - name: <name>
      files:
        - <file A>
        - path: "<file B>"
            - x: <x>
            - y: <y>
      x: <x>
      y: <y>
    - name: <name>
      folder: <folder>
  
theme:
    <cursor>: <name>
```