#!/usr/bin/env bash
# wt-go build script
#   ./Scripts/build.sh            -> mygo packages build/<os>-<arch>/wt-go (real testing uses this),
#                                    and the packages beside it are dropped: only the app stays
#   ./Scripts/build.sh --test     -> gofmt -l, go vet and go test before building
#   ./Scripts/build.sh --clean    -> delete build/ first
#   ./Scripts/build.sh --package  -> distribution only: go tool mygo build -o dist (no installer)
#   ./Scripts/build.sh --install  -> also put it where a Linux desktop keeps apps:
#                                    ~/.local/opt/wt-go and a menu entry, keeping data/
#   ./Scripts/build.sh --expect 0.3.0
#                                 -> assert mygo.json's version is that string (releases check the tag)
#
# The version has one source, mygo.json's version: mygo build writes it into the
# binary, and --expect only says which string the file must hold, so a release
# cannot be labelled one version and built from another.
#
# The icon, version info and native library are injected by mygo at package
# time (go build cannot carry resources), so a local build matches the
# release. The output directory comes from out in mygo.json (build/), and its
# name carries the platform so the upgrade manifest .mygo-files lines up with
# the artifacts: build/ and dist/ are in .gitignore.
set -euo pipefail

usage() {
    cat <<'EOF'
用法：
  ./Scripts/build.sh            构建到 build/<os>-<arch>/wt-go（真机测试就用它），
                                只留运行要的文件，.deb/.rpm/压缩包/安装脚本删掉
  ./Scripts/build.sh --test     编译前先跑 gofmt -l、go vet 与 go test
  ./Scripts/build.sh --clean    先删掉 build 目录再编译
  ./Scripts/build.sh --package  只出分发包：go tool mygo build -o dist（不造安装器）
  ./Scripts/build.sh --install  再装到本机（Linux 桌面）：~/.local/opt/wt-go 与菜单项，
                                升级不碰 data/（设置就存在里面）
  ./Scripts/build.sh --expect 0.3.0
                                断言 mygo.json 里的版本就是这一串（发版时对 tag 用）

平台取自 go env GOOS / GOARCH。
EOF
}

clean=0
test=0
package=0
install=0
expect=''
while [ "$#" -gt 0 ]; do
    case "$1" in
        --clean) clean=1 ;;
        --test) test=1 ;;
        --package) package=1 ;;
        --install) install=1 ;;
        --expect)
            if [ "$#" -lt 2 ]; then
                printf '  [ERR] --expect 后面要跟版本号（不带 v）。\n' >&2
                exit 2
            fi
            expect=$2
            shift
            ;;
        -h | --help)
            usage
            exit 0
            ;;
        *)
            printf '  [ERR] 不认识的参数：%s（用 --help 看用法）\n' "$1" >&2
            exit 2
            ;;
    esac
    shift
done

# The script sits in Scripts/; the repository is the directory above it, which
# go and mygo both run from.
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
root=$(dirname -- "$script_dir")
build_dir="$root/build"

if ! command -v go >/dev/null 2>&1; then
    printf '  [ERR] 找不到 go，请先安装 Go 1.27+。\n' >&2
    exit 1
fi

if [ ! -f "$root/mygo.json" ]; then
    printf '  [ERR] 找不到 %s：版本、图标与产物名都在它里面。\n' "$root/mygo.json" >&2
    exit 1
fi
# The version comes from mygo.json alone; --expect is only an assertion.
# awk, not sed | head: under pipefail a writer killed by SIGPIPE would take the
# script down with it.
target=$(awk -F'"' '/"version"[[:space:]]*:/ { print $4; exit }' "$root/mygo.json")
if ! printf '%s' "$target" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$'; then
    printf '  [ERR] mygo.json 的 version 要是不带 v 的三段数字（tag 去掉 v）：%s\n' "$target" >&2
    exit 1
fi
if [ -n "$expect" ] && [ "$expect" != "$target" ]; then
    printf '  [ERR] 版本对不上：--expect 是 %s，mygo.json 里是 %s。\n' "$expect" "$target" >&2
    exit 1
fi
# --package stops at dist/, so a box that asked for both would silently not
# install: say which flag it is answering.
if [ "$package" -eq 1 ] && [ "$install" -eq 1 ]; then
    printf '  [ERR] --package 只出分发包，--install 装到本机：只能挑一个。\n' >&2
    exit 2
fi

# So `wt-go version` names the release as well as the build: mygo appends this
# to its own -ldflags, hence no -ldflags= prefix here.
export MYGO_LDFLAGS="-X wtgo/internal/cli.packagedVersion=$target"

# Ask go instead of guessing, so macOS gets build/darwin-arm64/ without a flag.
goos=$(go env GOOS)
goarch=$(go env GOARCH)
platform="$goos/$goarch"
exe="$build_dir/$goos-$goarch/wt-go"

if [ "$clean" -eq 1 ] && [ -d "$build_dir" ]; then
    rm -rf -- "$build_dir"
fi

cd -- "$root"

if [ "$test" -eq 1 ]; then
    # CI's first gate is gofmt. Only the files git tracks: refs/ is read-only
    # reference material and would otherwise fail every local build.
    unformatted=$(git ls-files -z -- '*.go' | xargs -0 gofmt -l)
    if [ -n "$unformatted" ]; then
        printf '  [ERR] 需要 gofmt：\n%s\n' "$unformatted" >&2
        exit 1
    fi
    go vet ./... || {
        printf '  [ERR] go vet 失败。\n' >&2
        exit 1
    }
    # ./... rather than ./internal/...: the locales coverage tests run in CI too,
    # and a local build should not check less than a release does.
    go test ./... || {
        printf '  [ERR] 测试失败。\n' >&2
        exit 1
    }
fi

if [ "$package" -eq 1 ]; then
    go tool mygo build -platform "$platform" -skip-installer -skip-dmg -o dist || {
        printf '  [ERR] 打包失败。\n' >&2
        exit 1
    }
    printf '  [OK] %s v%s\n' "$root/dist" "$target"
    exit 0
fi

mkdir -p -- "$(dirname -- "$exe")"

# A running test instance holds the output file open, and Linux then refuses
# to write it (ETXTBSY, an exec of the truncated binary otherwise): find who
# holds it and say so. macOS has no /proc and allows replacing a running
# binary, so the check only applies where it can.
if [ -d /proc ]; then
    exe_real=$(cd -- "$(dirname -- "$exe")" && pwd)/$(basename -- "$exe")
    pids=''
    for link in /proc/[0-9]*/exe; do
        [ -L "$link" ] || continue
        # target_link, not target: target is the version, and readlink would
        # overwrite it with a path before the build reports it.
        target_link=$(readlink -- "$link" 2>/dev/null) || continue
        if [ "$target_link" = "$exe_real" ]; then
            pid=${link#/proc/}
            pid=${pid%/exe}
            pids="${pids:+$pids, }$pid"
        fi
    done
    if [ -n "$pids" ]; then
        printf '  [ERR] 正在运行的 wt-go（PID %s）占着输出文件：先关掉它的窗口再编译。\n' "$pids" >&2
        exit 1
    fi
fi

# mygo writes build/<os>-<arch>/ itself, per out in mygo.json; no staging
# directory, so the artifacts keep matching .mygo-files.
go tool mygo build -platform "$platform" -skip-installer -skip-dmg || {
    printf '  [ERR] 编译失败。\n' >&2
    exit 1
}

# What a package or an archive wants is a release's, and --package writes it
# into dist/; a local build keeps the app itself, as .mygo-files lists it. mygo
# writes the rest either way, so they are removed here rather than shipped
# beside the binary that a test run starts.
(
    cd -- "$(dirname -- "$exe")" || exit 0
    rm -f -- ./*.deb ./*.rpm ./*.tar.gz ./PKGBUILD ./install.sh
    for png in ./*-[0-9]*x[0-9]*.png; do
        [ -e "$png" ] && rm -f -- "$png"
    done
)

bytes=$(wc -c <"$exe" | tr -d ' ')
printf '  [OK] %s v%s (%d MB)\n' "$exe" "$target" "$((bytes / 1048576))"

# --install puts this build where a Linux desktop already keeps an app: one
# directory of files under ~/.local/opt, plus its menu entry.
if [ "$install" -eq 1 ]; then
    # macOS ships a signed .app and Windows an NSIS installer; neither is a
    # directory of files beside a binary, so this is a Linux desktop's layout.
    if [ "$goos" != linux ]; then
        printf '  [ERR] --install 只装 Linux 桌面（macOS 出 .app，Windows 出安装器）。\n' >&2
        exit 1
    fi
    src=$(dirname -- "$exe")
    app_dir="$HOME/.local/opt/wt-go"
    menu_dir="${XDG_DATA_HOME:-$HOME/.local/share}/applications"
    manifest="$src/.mygo-files"
    if [ ! -f "$manifest" ]; then
        printf '  [ERR] 找不到 %s：安装清单由 mygo 写在产物旁边。\n' "$manifest" >&2
        exit 1
    fi
    mkdir -p -- "$app_dir" "$menu_dir"
    # The manifest, not a file list kept here as well: mygo's own installer
    # copies the same set for an upgrade, so the two cannot drift apart.
    while IFS= read -r entry; do
        case "$entry" in
            # data/ is the settings directory, handled below; the menu entry is
            # rewritten with absolute paths, also below.
            '' | data | wt-go.desktop) continue ;;
        esac
        [ -e "$src/$entry" ] || continue
        # Rename over the old file instead of writing it: a running wt-go holds
        # the binary open and Linux refuses in-place writes (ETXTBSY), while
        # replacing the name is always allowed.
        cp -f -- "$src/$entry" "$app_dir/$entry.new"
        mv -f -- "$app_dir/$entry.new" "$app_dir/$entry"
    done <"$manifest"
    # data/ is what makes the copy portable, and settings.json lives in it:
    # create it when missing, never replace it, so an upgrade keeps settings.
    if [ ! -d "$app_dir/data" ]; then
        mkdir -p -- "$app_dir/data"
        if [ -f "$src/data/README.txt" ]; then
            cp -f -- "$src/data/README.txt" "$app_dir/data/README.txt"
        fi
    fi
    # The entry mygo writes names the binary by name, which only resolves from
    # inside its directory; the installed copy names the paths it installed.
    sed -e "s|^Exec=.*|Exec=\"$app_dir/wt-go\"|" \
        -e "s|^Icon=.*|Icon=$app_dir/icon.png|" \
        -- "$src/wt-go.desktop" >"$app_dir/wt-go.desktop"
    cp -f -- "$app_dir/wt-go.desktop" "$menu_dir/wt-go.desktop"
    # Menus cache their entries; without this the new one can wait for a login.
    if command -v update-desktop-database >/dev/null 2>&1; then
        update-desktop-database "$menu_dir" >/dev/null 2>&1 || true
    fi
    printf '  [OK] 已装 v%s 到 %s（菜单项 %s）\n' "$target" "$app_dir" "$menu_dir/wt-go.desktop"
fi