#!/usr/bin/env bash
# wt-go build script
#   ./Scripts/build.sh            -> mygo packages build/<os>-<arch>/wt-go (real testing uses this)
#   ./Scripts/build.sh --test     -> go vet and go test before building
#   ./Scripts/build.sh --clean    -> delete build/ first
#   ./Scripts/build.sh --package  -> distribution only: go tool mygo build -o dist (no installer)
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
  ./Scripts/build.sh            构建到 build/<os>-<arch>/wt-go（真机测试就用它）
  ./Scripts/build.sh --test     编译前先跑 go vet 与 go test
  ./Scripts/build.sh --clean    先删掉 build 目录再编译
  ./Scripts/build.sh --package  只出分发包：go tool mygo build -o dist（不造安装器）
  ./Scripts/build.sh --expect 0.3.0
                                断言 mygo.json 里的版本就是这一串（发版时对 tag 用）

平台取自 go env GOOS / GOARCH。
EOF
}

clean=0
test=0
package=0
expect=''
while [ "$#" -gt 0 ]; do
    case "$1" in
        --clean) clean=1 ;;
        --test) test=1 ;;
        --package) package=1 ;;
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

bytes=$(wc -c <"$exe" | tr -d ' ')
printf '  [OK] %s v%s (%d MB)\n' "$exe" "$target" "$((bytes / 1048576))"