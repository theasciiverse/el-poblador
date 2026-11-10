#!/usr/bin/env python3
"""Records asciiverse/el-poblador.cast, the catalog preview, at 100x30
(hub/RECORDING.md). It opens on the start screen, names four players and
starts; then a driver reads the game's help line from the screen and plays:
it walks the cursor over the island until a settlement spot is accepted,
lays the road beside it, and after the opening round rolls the dice and
ends each turn, moving the robber when a seven comes up. After about 44
seconds it quits with Ctrl+C.

The board is new every game, so nothing here depends on where things lie.
el_poblador runs by name from the folder given (default stage/), in an
empty HOME, so the cast's header holds no local path and no saves of your
own come into it.

Needs asciinema 3.x and pyte. Usage, from the repo root:
  uv run --with pyte python3 asciiverse/record-preview.py [bindir]
With --dry-run it plays without recording."""
import fcntl, os, pty, random, select, struct, sys, tempfile, termios, time
import pyte

root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
args = [a for a in sys.argv[1:] if a != "--dry-run"]
dry = "--dry-run" in sys.argv
bindir = os.path.abspath(args[0]) if args else f"{root}/stage"
cast = f"{root}/asciiverse/el-poblador.cast"
COLS, ROWS, PLAY = 100, 30, float(os.environ.get("PLAY", "44"))
KEYS = {"up": b"\x1b[A", "down": b"\x1b[B", "right": b"\x1b[C", "left": b"\x1b[D",
        "enter": b"\r", "tab": b"\t"}

home = tempfile.mkdtemp()
env = dict(os.environ, TERM="xterm-256color", COLORTERM="truecolor", LANG="C.UTF-8",
           HOME=home, USER="player", LOGNAME="player", XDG_STATE_HOME=f"{home}/.local/state",
           PATH=f"{bindir}:{os.environ['PATH']}")
for v in ("VTE_VERSION", "TERM_PROGRAM", "TERM_PROGRAM_VERSION", "TMUX", "COLORFGBG"):
    env.pop(v, None)

pid, fd = pty.fork()
if pid == 0:
    os.chdir(home)
    if dry:
        os.execvpe("el_poblador", ["el_poblador"], env)
    os.execvpe("asciinema", ["asciinema", "rec", "--overwrite", "-q", "--idle-time-limit", "2",
        "--window-size", f"{COLS}x{ROWS}", "--capture-env", "TERM,COLORTERM,LANG",
        "--title", "El Poblador", "--command", "el_poblador", cast], env)
fcntl.ioctl(fd, termios.TIOCSWINSZ, struct.pack("HHHH", ROWS, COLS, 0, 0))
screen = pyte.Screen(COLS, ROWS)
stream = pyte.ByteStream(screen)


def pump(seconds):
    end = time.time() + seconds
    while True:
        left = end - time.time()
        r, _, _ = select.select([fd], [], [], max(0.0, min(0.02, left)))
        if r:
            try:
                stream.feed(os.read(fd, 65536))
            except OSError:
                return False
        if left <= 0:
            return True


def key(k, wait=0.25):
    os.write(fd, KEYS.get(k, k.encode() if isinstance(k, str) else k))
    pump(wait)


def typed(text, wait=0.12):
    for ch in text:
        key(ch, wait)


def help_line():
    """The game's prompt: the last line on screen with text on it."""
    for line in reversed(screen.display):
        if line.strip():
            return line.strip()
    return ""


# Bubble Tea asks the terminal about itself at start; nothing answers in
# this pty, so the first frame can take a while (the idle limit squeezes it)
pump(1.5)
for _ in range(100):
    if "New game" in "\n".join(screen.display):
        break
    pump(0.2)
pump(1.2)
key("enter", 0.6)
for i, name in enumerate(["Ana", "Bruno", "Carmen", "Diego"]):
    if i == 3:  # on Start, right turns three players into four
        key("right", 0.5)
        key("up", 0.3)
    typed(name)
    key("tab", 0.3)
pump(0.8)
key("enter", 1.2)

start = time.time()
walk = ["up", "down", "left", "right"]
stuck = 0
while time.time() - start < PLAY:
    h = help_line()
    if "settlement" in h:
        # walk two or three steps, try the spot; refused spots keep the prompt
        # (a refused spot keeps the prompt, so the next try comes quickly)
        for _ in range(random.randint(1, 2) if stuck else 2):
            key(random.choice(walk), 0.12 if stuck else 0.22)
        before = h
        key("enter", 0.3)
        if help_line() == before:
            stuck += 1
        else:
            stuck = 0
    elif "road" in h:
        key(random.choice(walk), 0.3)
        key("enter", 0.6)
    elif "roll the dice" in h:
        pump(0.6)
        key("enter", 1.6)  # Roll is the first choice
    elif "robber" in h.lower():
        for _ in range(random.randint(1, 3)):
            key(random.choice(walk), 0.25)
        key("enter", 0.8)
    elif "What do you want to do" in h or "Anything else" in h:
        for _ in range(4):  # down to End Turn
            key("down", 0.2)
        pump(0.4)
        key("enter", 1.0)
    else:
        # a choice this driver doesn't know (stealing, discarding): take the first
        key("enter", 0.8)
    if stuck > 80:
        break

pump(1.5)
key(b"\x03", 1.0)
for _ in range(50):
    if os.waitpid(pid, os.WNOHANG)[0]:
        break
    pump(0.2)
