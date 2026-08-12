#!/usr/bin/env python3
"""Generate a printable work schedule PDF, blank or filled in.

Two ways in:

    build_schedule.py "Juan Perez" 06082026
        One blank sheet, every cell empty, for filling in by hand.

    build_schedule.py --json payload.json --outdir out/
    ... | build_schedule.py --json - --outdir out/
        One sheet per worker with their punches already filled in. The
        payload is what klapp's printsched binary sends to the schedule
        listener; see internal/schedule/payload.go for its definition.

Every row of the sheet can carry values: the fourteen day rows, the TOTAL
row, and the spare rows underneath it (a day with a second punch spills
into those, and the table grows if there are more than four).
"""

import argparse
import json
import os
import re
import sys
from reportlab.lib.pagesizes import letter
from reportlab.lib import colors
from reportlab.lib.units import inch
from reportlab.platypus import BaseDocTemplate, PageTemplate, Frame, Table, TableStyle, Paragraph
from reportlab.lib.styles import ParagraphStyle
from datetime import date, datetime, timedelta

# ── Configuration ────────────────────────────────────────────────────────────
WEEKS = 2
# The sheet always reserves at least this many spare rows under TOTAL.
MIN_BLANK_ROWS = 4
# ─────────────────────────────────────────────────────────────────────────────

BLACK = colors.black
WHITE = colors.white


def parse_iso(value):
    """Parse a "2006-01-02" string, or return None for empty/garbage."""
    if not value:
        return None
    try:
        return datetime.strptime(value, "%Y-%m-%d").date()
    except ValueError:
        return None


def date_label(d):
    """Render a date the way the DATE column wants it."""
    return f"{d.strftime('%A').upper()}   {d.strftime('%B %-d, %Y')}"


def slug(name):
    """Turn a worker name into something safe for a filename."""
    cleaned = re.sub(r"[^A-Za-z0-9]+", "_", name).strip("_")
    return cleaned or "schedule"


def row_cells(row, fallback_date=None):
    """Turn one payload row into the sheet's five columns.

    A row is {"date", "in", "out", "hours"} and every field may be empty,
    which prints an empty cell. fallback_date labels a day row whose
    payload entry carried no date of its own.
    """
    row = row or {}
    d = parse_iso(row.get("date")) or fallback_date
    label = date_label(d) if d else ""
    return [label, row.get("in", ""), row.get("out", ""), row.get("hours", ""), ""]


def build_schedule(name, start_date, days=None, extra=None, total="", outdir="."):
    """Write one worker's sheet and return the path it was written to.

    days is a list of up to WEEKS*7 payload rows, one per day in order;
    extra is the overflow printed under TOTAL; total is the TOTAL cell.
    All three default to empty, which produces the blank sheet.
    """
    days = days or []
    extra = extra or []

    output_file = os.path.join(outdir, f"{slug(name)}_{start_date.strftime('%m%d%Y')}_schedule.pdf")
    margin = 0.28 * inch
    doc = BaseDocTemplate(output_file, pagesize=letter,
                          leftMargin=margin, rightMargin=margin,
                          topMargin=margin, bottomMargin=margin)
    frame = Frame(margin, margin, letter[0] - 2*margin, letter[1] - 2*margin,
                  leftPadding=0, rightPadding=0, topPadding=0, bottomPadding=0)
    doc.addPageTemplates([PageTemplate(id='main', frames=[frame])])

    usable_w = letter[0] - 2 * margin
    usable_h = letter[1] - 2 * margin

    name_style = ParagraphStyle(
        "Name",
        fontName="Helvetica-Bold",
        fontSize=16,
        textColor=BLACK,
        leading=19,
    )
    sub_style = ParagraphStyle(
        "Sub",
        fontName="Helvetica",
        fontSize=9,
        textColor=BLACK,
        leading=12,
    )

    dates = [start_date + timedelta(days=i) for i in range(WEEKS * 7)]
    end_date = dates[-1]
    header_text = f"{start_date.strftime('%B %-d')} – {end_date.strftime('%B %-d, %Y')}"

    n_data_rows = len(dates)
    # The spare rows are where a day's second punch goes, so the table has
    # to grow when someone punched in and out more than once in a day.
    n_blank_rows = max(MIN_BLANK_ROWS, len(extra))

    # Put everything in one table so heights are exact — no inter-flowable gaps
    # Rows: name, date-range, column headers, 14 days, TOTAL
    name_row_h   = 0.38 * inch
    col_hdr_row_h = 0.30 * inch
    total_row_h  = 0.30 * inch
    fixed_h      = name_row_h + col_hdr_row_h + total_row_h
    data_row_h   = (usable_h * 0.5 - fixed_h) / (n_data_rows + n_blank_rows)

    date_col_w  = usable_w * 0.38
    other_col_w = (usable_w - date_col_w) / 5
    extra_col_w = other_col_w * 2
    col_widths  = [date_col_w, other_col_w, other_col_w, other_col_w, extra_col_w]

    # row indices
    R_NAME        = 0
    R_HDR         = 1
    R_FIRST       = 2
    R_LAST        = R_FIRST + n_data_rows - 1
    R_TOTAL       = R_LAST + 1
    R_BLANK_FIRST = R_TOTAL + 1
    R_BLANK_LAST  = R_BLANK_FIRST + n_blank_rows - 1

    table_data = [
        [name, "", "", "", ""],
        ["DATE", "ENTRADA", "SALIDA", "HORAS", ""],
    ]
    for i, d in enumerate(dates):
        row = days[i] if i < len(days) else None
        table_data.append(row_cells(row, fallback_date=d))
    table_data.append(["TOTAL:", "", "", total, ""])
    for i in range(n_blank_rows):
        row = extra[i] if i < len(extra) else None
        table_data.append(row_cells(row))

    row_heights = (
        [name_row_h, col_hdr_row_h]
        + [data_row_h] * n_data_rows
        + [total_row_h]
        + [data_row_h] * n_blank_rows
    )

    tbl = Table(table_data, colWidths=col_widths, rowHeights=row_heights)

    style_cmds = [
        # ── Name row ────────────────────────────────────────────────────────
        ("SPAN",        (0, R_NAME),  (-1, R_NAME)),
        ("FONTNAME",    (0, R_NAME),  (-1, R_NAME),  "Helvetica-Bold"),
        ("FONTSIZE",    (0, R_NAME),  (-1, R_NAME),  16),
        ("VALIGN",      (0, R_NAME),  (-1, R_NAME),  "MIDDLE"),
        ("ALIGN",       (0, R_NAME),  (-1, R_NAME),  "LEFT"),
        ("LEFTPADDING", (0, R_NAME),  (-1, R_NAME),  4),
        ("LINEBEFORE",  (0, R_NAME),  (0, R_NAME),   0, BLACK),
        ("LINEAFTER",   (0, R_NAME),  (-1, R_NAME),  0, BLACK),
        ("LINEABOVE",   (0, R_NAME),  (-1, R_NAME),  0, BLACK),

        # ── Column header row ───────────────────────────────────────────────
        ("FONTNAME",    (0, R_HDR),   (-1, R_HDR),   "Helvetica-Bold"),
        ("FONTSIZE",    (0, R_HDR),   (-1, R_HDR),   10),
        ("ALIGN",       (0, R_HDR),   (-1, R_HDR),   "CENTER"),
        ("VALIGN",      (0, R_HDR),   (-1, R_HDR),   "MIDDLE"),
        ("LINEABOVE",   (0, R_HDR),   (-1, R_HDR),   1.5, BLACK),
        ("LINEBELOW",   (0, R_HDR),   (-1, R_HDR),   1.2, BLACK),

        # ── Data rows ───────────────────────────────────────────────────────
        ("FONTNAME",    (0, R_FIRST), (-1, R_LAST),  "Helvetica"),
        ("FONTSIZE",    (0, R_FIRST), (-1, R_LAST),  8.5),
        ("ALIGN",       (1, R_FIRST), (-1, R_LAST),  "CENTER"),
        ("VALIGN",      (0, R_FIRST), (-1, R_LAST),  "MIDDLE"),
        ("LEFTPADDING", (0, R_FIRST), (0, R_LAST),   8),
        ("FONTNAME",    (0, R_FIRST), (0, R_LAST),   "Helvetica-Bold"),
        ("FONTSIZE",    (0, R_FIRST), (0, R_LAST),   8.7),

        # ── TOTAL row ───────────────────────────────────────────────────────
        ("FONTNAME",    (0, R_TOTAL), (-1, R_TOTAL), "Helvetica-Bold"),
        ("FONTSIZE",    (0, R_TOTAL), (-1, R_TOTAL), 9),
        ("ALIGN",       (0, R_TOTAL), (-1, R_TOTAL), "CENTER"),
        ("VALIGN",      (0, R_TOTAL), (-1, R_TOTAL), "MIDDLE"),
        ("LINEABOVE",   (0, R_TOTAL), (-1, R_TOTAL), 1.2, BLACK),

        # ── Blank rows ──────────────────────────────────────────────────────
        ("FONTNAME",    (0, R_BLANK_FIRST), (-1, R_BLANK_LAST), "Helvetica"),
        ("FONTSIZE",    (0, R_BLANK_FIRST), (-1, R_BLANK_LAST), 8.5),
        ("ALIGN",       (1, R_BLANK_FIRST), (-1, R_BLANK_LAST), "CENTER"),
        ("VALIGN",      (0, R_BLANK_FIRST), (-1, R_BLANK_LAST), "MIDDLE"),
        ("LEFTPADDING", (0, R_BLANK_FIRST), (0, R_BLANK_LAST),  8),
        # A spilled punch keeps the DATE column's bold, same as the day rows.
        ("FONTNAME",    (0, R_BLANK_FIRST), (0, R_BLANK_LAST),  "Helvetica-Bold"),
        ("FONTSIZE",    (0, R_BLANK_FIRST), (0, R_BLANK_LAST),  8.7),

        # ── Grid (data + total + blank rows) ─────────────────────────────────
        ("GRID",        (0, R_HDR),   (-1, R_BLANK_LAST), 0.5, BLACK),
        ("BOX",         (0, R_HDR),   (-1, R_BLANK_LAST), 1,   BLACK),
    ]

    tbl.setStyle(TableStyle(style_cmds))

    doc.build([tbl])
    return output_file


def build_from_payload(payload, outdir="."):
    """Build every sheet in a print job. Returns the paths written."""
    start = parse_iso(payload.get("start_date") or payload.get("pay_period"))
    if start is None:
        raise ValueError("payload has no usable start_date or pay_period")

    os.makedirs(outdir, exist_ok=True)

    written = []
    for sheet in payload.get("sheets") or []:
        written.append(build_schedule(
            sheet.get("name", ""),
            start,
            days=sheet.get("days"),
            extra=sheet.get("extra"),
            total=sheet.get("total", ""),
            outdir=outdir,
        ))
    return written


def main():
    parser = argparse.ArgumentParser(description="Generate a work schedule PDF.")
    parser.add_argument("name", nargs="?", help="Employee name (blank-sheet mode)")
    parser.add_argument("startdate", nargs="?", help="Start date in mmddyyyy format (e.g. 06082026)")
    parser.add_argument("--json", dest="json_path",
                        help="Path to a print-job JSON payload, or - for stdin")
    parser.add_argument("--outdir", default=".", help="Directory to write PDFs into")
    args = parser.parse_args()

    if args.json_path:
        raw = sys.stdin.read() if args.json_path == "-" else open(args.json_path).read()
        written = build_from_payload(json.loads(raw), args.outdir)
        if not written:
            print("No sheets in payload", file=sys.stderr)
            return 1
        for path in written:
            print(path)
        return 0

    if not args.name or not args.startdate:
        parser.error("name and startdate are required unless --json is given")

    os.makedirs(args.outdir, exist_ok=True)
    start = date(int(args.startdate[4:8]), int(args.startdate[0:2]), int(args.startdate[2:4]))
    print(build_schedule(args.name, start, outdir=args.outdir))
    return 0


if __name__ == "__main__":
    sys.exit(main())
