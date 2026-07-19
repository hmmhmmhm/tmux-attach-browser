Output docs/assets/tab-demo.gif

Require go
Require tmux

Set Shell "bash"
Set FontSize 14
Set Width 800
Set Height 500
Set Framerate 10
Set TypingSpeed 60ms
Set Theme "Catppuccin Mocha"
Set Padding 12
Set Margin 0
Set WindowBar Rings
Set CursorBlink false

Hide
Type `unset TMUX`
Enter
Type `export TAB_DEMO_DIR="$(mktemp -d /tmp/tab-demo.XXXXXX)"`
Enter
Type `export TMUX_TMPDIR="$TAB_DEMO_DIR/tmux"`
Enter
Type `mkdir -p "$TMUX_TMPDIR" "$TAB_DEMO_DIR/bin"`
Enter
Type `go build -o "$TAB_DEMO_DIR/bin/tab" ./cmd/tab`
Enter
Type `export PATH="$TAB_DEMO_DIR/bin:$PATH"`
Enter
Type `export PS1='$ '`
Enter
Type `tmux -f /dev/null new-session -d -s bootstrap -n shell`
Enter
Type `tmux set-option -g default-shell /bin/bash`
Enter
Type `tmux set-option -g default-command "exec /bin/bash --noprofile --norc -i"`
Enter
Type `tmux set-option -g status-left " session: #S "`
Enter
Type `tmux set-option -g status-left-length 20`
Enter
Type `tmux set-option -g status-right ""`
Enter
Type `tmux set-option -g status-style "bg=#313244,fg=#cdd6f4"`
Enter
Type `tmux new-session -d -s work -n editor`
Enter
Type `tmux new-window -d -t work -n server`
Enter
Type `tmux new-window -d -t work -n logs`
Enter
Type `sleep 1`
Enter
Type `tmux new-session -d -s docs -n writing`
Enter
Type `sleep 1`
Enter
Type `tmux new-session -d -s api -n shell`
Enter
Type `tmux new-window -d -t api -n tests`
Enter
Type `tmux send-keys -t api:0.0 "export PS1='$ '; clear" C-m`
Enter
Type `tmux kill-session -t bootstrap`
Enter
Type `clear`
Enter
Show

Type "tab"
Sleep 300ms
Enter
Wait+Screen /tmux sessions/
Sleep 1s
Down
Sleep 600ms
Up
Sleep 600ms
Enter
Sleep 1500ms

Type "tab"
Sleep 300ms
Enter
Wait+Screen /tmux sessions/
Sleep 1s
Type "n"
Wait+Screen /session name:/
Sleep 500ms
Type "demo"
Sleep 500ms
Enter
Wait+Screen /session: demo/
Sleep 3s

Hide
Ctrl+B
Type "d"
Sleep 1s
Type `tmux kill-server`
Enter
