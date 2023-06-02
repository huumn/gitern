package art

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Art []string

var Magic Art = Art{
	`        .      _   `,
	`        |\    | |  `,
	`       _/_\_  0X0  `,
	`      ,_\-/_   |   `,
	`     /  | | \_ |   `,
	`     \ \| |\__(/   `,
	`     /\/   \   |   `,
	`    /       \  |   `,
	` _.'         \ |   `,
	` '------.--.-' |   `,
}

var Scroll Art = Art{
	`  _____________   `,
	` (_\           \  `,
	`   |           |  `,
	`   |           |  `,
	`   |  granted  |  `,
	`   |           |  `,
	`   |___________|  `,
	`  @)____________) `,
}

var Skull Art = Art{
	`       .-----.       `,
	`      /,     ,\      `,
	`     | )() ()( |     `,
	` {}_ (/_  ^  _\) _{} `,
	`    "=| IIIII |="    `,
	`    _.=\_____/=._    `,
	` {}"             "{} `,
}

var Scales Art = Art{
	`           II            `,
	`    .______()______.     `,
	`   /|\     ||     /|\    `,
	`  / | \    ||    / | \   `,
	` (__|__)   ||   (__|__)  `,
	`           ||            `,
	`           ||            `,
	`         __||__          `,
	`        /______\         `,
}

var Fool Art = Art{
	`       __       `,
	`  _   //\( __   `,
	" //`\\/ | @/ \\\\  ",
	` ||. \_|_/ .||  `,
	` )/|_______|\(  `,
	` @ | ^   ^ | @  `,
	`   \  o  o /    `,
	`   (    @  )    `,
	"    \\ `--'/     ",
	"     `---'      ",
}

var Noose Art = Art{
	`     ||      `,
	`     ||      `,
	`    (-_)     `,
	`    (_-)     `,
	`   //  \\    `,
	`  //    \\   `,
	` //      \\  `,
	` ||      ||  `,
	` \\      //  `,
	"  ``====''   ",
}

var Sun Art = Art{
	`         .            `,
	`   '.    |    .'      `,
	`_    \ .-''-./     _  `,
	` "-._/        \_.-"   `,
	`- - |          | - -  `,
	`  _."\        /"._    `,
	`-"   / '-..-' \   "-  `,
	"   .'     |    `.     ",
	`          '           `,
}

var Tower Art = Art{
	`     |        `,
	`     A        `,
	`    / \       `,
	`    |||  \/   `,
	`    |||   ^.  `,
	`    |||       `,
	`    |||.(     `,
	`  .(||/\%\    `,
	` /%/\(%(%))   `,
}

func (art Art) Fprint(w io.Writer, msg ...string) {
	nlines := len(art)
	if len(msg) > nlines {
		nlines = len(msg)
	} else {
		// take the difference, divide it by two, that's how many blank lines to
		// prepend to msg
		msg = append(make([]string, (nlines-len(msg))/2), msg...)
	}

	for i := 0; i < nlines; i++ {
		var prefix, message string
		if i < len(art) {
			prefix = art[i]
		} else {
			prefix = strings.Repeat(" ", len(art[0]))
		}
		if i < len(msg) {
			message = msg[i]
		}
		fmt.Fprintln(w, prefix, message)
	}
}

func (art Art) Print(msg ...string) {
	art.Fprint(os.Stdout, msg...)
}

func (art Art) Fatal(msg ...string) {
	art.Fprint(os.Stderr, msg...)
	os.Exit(1)
}
