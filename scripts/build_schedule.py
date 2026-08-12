#!/usr/bin/env python3
"""Generate a printable work schedule PDF."""

import argparse
from reportlab.lib.pagesizes import letter
from reportlab.lib import colors
from reportlab.lib.units import inch
from reportlab.platypus import BaseDocTemplate, PageTemplate, Frame, Table, TableStyle, Paragraph
from reportlab.lib.styles import ParagraphStyle
from datetime import date, timedelta

# ── Configuration ────────────────────────────────────────────────────────────
WEEKS = 2
# ─────────────────────────────────────────────────────────────────────────────

BLACK = colors.black
WHITE = colors.white


def build_schedule(name, start_date):
    output_file = f"{name}_{start_date.strftime('%m%d%Y')}_schedule.pdf"
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

    # Put everything in one table so heights are exact — no inter-flowable gaps
    # Rows: name, date-range, column headers, 14 days, TOTAL
    name_row_h   = 0.38 * inch
    col_hdr_row_h = 0.30 * inch
    total_row_h  = 0.30 * inch
    fixed_h      = name_row_h + col_hdr_row_h + total_row_h
    data_row_h   = (usable_h * 0.5 - fixed_h) / (len(dates) + 4)

    date_col_w  = usable_w * 0.38
    other_col_w = (usable_w - date_col_w) / 5
    extra_col_w = other_col_w * 2
    col_widths  = [date_col_w, other_col_w, other_col_w, other_col_w, extra_col_w]

    n_data_rows  = len(dates)
    n_blank_rows = 4
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
    for d in dates:
        date_str = f"{d.strftime('%A').upper()}   {d.strftime('%B %-d, %Y')}"
        table_data.append([date_str, "", "", "", ""])
    table_data.append(["TOTAL:", "", "", "", ""])
    for _ in range(n_blank_rows):
        table_data.append(["", "", "", "", ""])

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

        # ── Grid (data + total + blank rows) ─────────────────────────────────
        ("GRID",        (0, R_HDR),   (-1, R_BLANK_LAST), 0.5, BLACK),
        ("BOX",         (0, R_HDR),   (-1, R_BLANK_LAST), 1,   BLACK),
    ]

    tbl.setStyle(TableStyle(style_cmds))

    doc.build([tbl])
    print(f"Saved → {output_file}")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate a work schedule PDF.")
    parser.add_argument("name", help="Employee name")
    parser.add_argument("startdate", help="Start date in mmddyyyy format (e.g. 06082026)")
    args = parser.parse_args()

    start = date(int(args.startdate[4:8]), int(args.startdate[0:2]), int(args.startdate[2:4]))
    build_schedule(args.name, start)
