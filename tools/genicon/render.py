# /// script
# dependencies = ["resvg-py==0.5.0"]
# ///
import argparse
from pathlib import Path

from resvg_py import svg_to_bytes


def main() -> None:
    resources = Path(__file__).resolve().parents[2] / "resources"
    parser = argparse.ArgumentParser(description="Render the wt-go application icon.")
    parser.add_argument("output", nargs="?", type=Path, default=resources / "icon.png")
    output = parser.parse_args().output
    png = svg_to_bytes(
        svg_path=str(resources / "icon.svg"),
        width=1024,
        height=1024,
        skip_system_fonts=True,
    )
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_bytes(png)


if __name__ == "__main__":
    main()
