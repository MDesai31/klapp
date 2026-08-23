#!/usr/bin/env python3
"""Tests for build_schedule.py.

    ~/venv/bin/python3 -m unittest discover -s schedule_listener

stdlib unittest, no pytest, because the only thing this box's venv has to
carry for the listener is reportlab.

Two layers: sheet_rows() decides what lands in which cell and is checked
directly, and the PDFs are built for real and read back with pdftotext
(skipped when poppler is missing) so a payload that draws nothing on the
page can't pass.
"""

import json
import os
import shutil
import subprocess
import sys
import tempfile
import unittest
from datetime import date

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import build_schedule as bs

START = date(2026, 6, 8)
SCRIPT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "build_schedule.py")

# Row indices into sheet_rows()'s table: name, column headers, then the
# fourteen day rows, TOTAL, and the spare rows.
R_HDR = 1
R_FIRST_DAY = 2
R_TOTAL = R_FIRST_DAY + 14
R_FIRST_SPARE = R_TOTAL + 1


def day_row(d, tin="", tout="", hours=""):
    return {"date": d, "in": tin, "out": tout, "hours": hours}


def full_period(overrides=None):
    """Fourteen empty day rows, June 8-21 2026, with some filled in.

    overrides maps a day index to the row that replaces the empty one.
    """
    days = [day_row(f"2026-06-{8 + i:02d}") for i in range(14)]
    for i, row in (overrides or {}).items():
        days[i] = row
    return days


def pdftotext(path):
    """The PDF's text as one string, laid out left to right.

    pdftotext measures gaps rather than copying them, so the DATE column's
    "MONDAY   June 8" comes back single-spaced; assertions here allow for
    that instead of pinning the exact run of spaces.
    """
    return subprocess.run(
        ["pdftotext", "-layout", path, "-"],
        check=True, capture_output=True, text=True,
    ).stdout


have_pdftotext = shutil.which("pdftotext") is not None


class HelperTests(unittest.TestCase):
    def test_parse_iso(self):
        self.assertEqual(bs.parse_iso("2026-06-08"), date(2026, 6, 8))
        for bad in ("", None, "06/08/2026", "2026-13-01", "garbage"):
            self.assertIsNone(bs.parse_iso(bad), bad)

    def test_date_label(self):
        self.assertEqual(bs.date_label(date(2026, 6, 8)), "MONDAY   June 8, 2026")

    def test_slug(self):
        self.assertEqual(bs.slug("Juan Perez"), "Juan_Perez")
        self.assertEqual(bs.slug("José Ramírez-Díaz"), "Jos_Ram_rez_D_az")
        self.assertEqual(bs.slug("../../etc/passwd"), "etc_passwd")
        self.assertEqual(bs.slug(""), "schedule")
        self.assertEqual(bs.slug("!!!"), "schedule")

    def test_row_cells_fills_every_column(self):
        cells = bs.row_cells(day_row("2026-06-08", "7:00 AM", "3:30 PM", "8:30"))
        self.assertEqual(cells, ["MONDAY   June 8, 2026", "7:00 AM", "3:30 PM", "8:30", ""])

    def test_row_cells_falls_back_to_the_day_it_labels(self):
        cells = bs.row_cells(None, fallback_date=date(2026, 6, 9))
        self.assertEqual(cells, ["TUESDAY   June 9, 2026", "", "", "", ""])

    def test_row_cells_of_nothing_is_all_empty(self):
        self.assertEqual(bs.row_cells(None), ["", "", "", "", ""])


class SheetRowTests(unittest.TestCase):
    def test_blank_sheet_shape(self):
        data, n_days, n_spare = bs.sheet_rows("Juan Perez", START)

        self.assertEqual((n_days, n_spare), (14, bs.MIN_BLANK_ROWS))
        # name + header + 14 days + TOTAL + 4 spares
        self.assertEqual(len(data), 21)
        self.assertEqual(data[0][0], "Juan Perez")
        self.assertEqual(data[R_HDR], ["DATE", "ENTRADA", "SALIDA", "HORAS", ""])
        self.assertEqual(data[R_TOTAL], ["TOTAL:", "", "", "", ""])
        # Every day row is labelled and otherwise empty, ready for a pen.
        for i in range(14):
            row = data[R_FIRST_DAY + i]
            self.assertEqual(row[1:], ["", "", "", ""])
            self.assertTrue(row[0])

    def test_fourteen_days_are_consecutive(self):
        data, _, _ = bs.sheet_rows("Juan Perez", START)
        labels = [data[R_FIRST_DAY + i][0] for i in range(14)]
        self.assertEqual(labels[0], "MONDAY   June 8, 2026")
        self.assertEqual(labels[-1], "SUNDAY   June 21, 2026")

    def test_punches_land_on_their_own_day(self):
        days = full_period({
            0: day_row("2026-06-08", "7:00 AM", "3:30 PM", "8:30"),
            13: day_row("2026-06-21", "8:00 AM", "12:00 PM", "4:00"),
        })
        data, _, _ = bs.sheet_rows("Juan Perez", START, days=days, total="12:30")

        self.assertEqual(data[R_FIRST_DAY][1:4], ["7:00 AM", "3:30 PM", "8:30"])
        self.assertEqual(data[R_FIRST_DAY + 13][1:4], ["8:00 AM", "12:00 PM", "4:00"])
        self.assertEqual(data[R_FIRST_DAY + 1][1:4], ["", "", ""])
        self.assertEqual(data[R_TOTAL][3], "12:30")

    def test_open_punch_prints_in_without_out_or_hours(self):
        days = full_period({2: day_row("2026-06-10", "6:45 AM")})
        data, _, _ = bs.sheet_rows("Juan Perez", START, days=days)
        self.assertEqual(data[R_FIRST_DAY + 2][1:4], ["6:45 AM", "", ""])

    def test_spare_rows_carry_the_spill(self):
        extra = [day_row("2026-06-08", "5:00 PM", "8:00 PM", "3:00")]
        data, _, n_spare = bs.sheet_rows("Juan Perez", START, extra=extra)

        self.assertEqual(n_spare, bs.MIN_BLANK_ROWS)
        self.assertEqual(
            data[R_FIRST_SPARE],
            ["MONDAY   June 8, 2026", "5:00 PM", "8:00 PM", "3:00", ""],
        )
        # The unused spares stay blank.
        self.assertEqual(data[R_FIRST_SPARE + 1], ["", "", "", "", ""])

    def test_table_grows_past_four_spares(self):
        extra = [day_row(f"2026-06-{8 + i:02d}", "5:00 PM", "8:00 PM", "3:00") for i in range(7)]
        data, _, n_spare = bs.sheet_rows("Juan Perez", START, extra=extra)

        self.assertEqual(n_spare, 7)
        self.assertEqual(len(data), 24)
        for i in range(7):
            self.assertEqual(data[R_FIRST_SPARE + i][1:4], ["5:00 PM", "8:00 PM", "3:00"])

    def test_short_days_list_still_labels_all_fourteen(self):
        # A payload that only sent a few rows must not shorten the sheet.
        data, n_days, _ = bs.sheet_rows("Juan Perez", START, days=full_period()[:3])
        self.assertEqual(n_days, 14)
        self.assertEqual(data[R_FIRST_DAY + 13][0], "SUNDAY   June 21, 2026")


class BuildPDFTests(unittest.TestCase):
    def setUp(self):
        self.outdir = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, self.outdir)

    def test_blank_sheet_is_a_pdf_named_for_the_worker(self):
        path = bs.build_schedule("Juan Perez", START, outdir=self.outdir)

        self.assertEqual(os.path.basename(path), "Juan_Perez_06082026_schedule.pdf")
        with open(path, "rb") as f:
            self.assertEqual(f.read(5), b"%PDF-")
        self.assertGreater(os.path.getsize(path), 1000)

    @unittest.skipUnless(have_pdftotext, "poppler's pdftotext is not installed")
    def test_filled_sheet_puts_the_hours_on_the_page(self):
        days = full_period({0: day_row("2026-06-08", "7:00 AM", "3:30 PM", "8:30")})
        extra = [day_row("2026-06-08", "5:00 PM", "8:00 PM", "3:00")]
        path = bs.build_schedule("Maria Lopez", START, days=days, extra=extra,
                                 total="11:30", outdir=self.outdir)

        text = pdftotext(path)
        self.assertIn("Maria Lopez", text)
        self.assertIn("ENTRADA", text)
        self.assertIn("MONDAY June 8, 2026", text)
        self.assertIn("7:00 AM", text)
        self.assertIn("3:30 PM", text)
        self.assertIn("5:00 PM", text)   # the spill row
        self.assertIn("11:30", text)     # TOTAL
        self.assertIn("SUNDAY June 21, 2026", text)

    @unittest.skipUnless(have_pdftotext, "poppler's pdftotext is not installed")
    def test_a_full_sheet_still_fits_one_page(self):
        days = full_period({
            i: day_row(f"2026-06-{8 + i:02d}", "7:00 AM", "3:30 PM", "8:30") for i in range(14)
        })
        extra = [day_row(f"2026-06-{8 + i:02d}", "5:00 PM", "8:00 PM", "3:00") for i in range(10)]
        path = bs.build_schedule("Juan Perez", START, days=days, extra=extra,
                                 total="149:00", outdir=self.outdir)

        self.assertEqual(pdftotext(path).count("\f"), 1)


class PayloadTests(unittest.TestCase):
    """The payload half: what the listener actually hands to this script."""

    def setUp(self):
        self.outdir = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, self.outdir)

    def payload(self, **overrides):
        p = {
            "pay_period": "2026-06-08",
            "start_date": "2026-06-08",
            "end_date": "2026-06-21",
            "sheets": [
                {
                    "name": "Juan Perez",
                    "days": full_period({0: day_row("2026-06-08", "7:00 AM", "3:30 PM", "8:30")}),
                    "extra": [],
                    "total": "8:30",
                },
                {"name": "Maria Lopez", "days": full_period(), "extra": [], "total": ""},
            ],
        }
        p.update(overrides)
        return p

    def test_one_pdf_per_sheet(self):
        written = bs.build_from_payload(self.payload(), self.outdir)

        self.assertEqual([os.path.basename(p) for p in written], [
            "Juan_Perez_06082026_schedule.pdf",
            "Maria_Lopez_06082026_schedule.pdf",
        ])
        for path in written:
            self.assertTrue(os.path.exists(path))

    def test_outdir_is_created(self):
        nested = os.path.join(self.outdir, "a", "b")
        written = bs.build_from_payload(self.payload(), nested)
        self.assertTrue(os.path.exists(written[0]))

    def test_pay_period_stands_in_for_a_missing_start_date(self):
        p = self.payload()
        del p["start_date"]
        written = bs.build_from_payload(p, self.outdir)
        self.assertIn("06082026", written[0])

    def test_undated_payload_is_an_error(self):
        with self.assertRaises(ValueError):
            bs.build_from_payload({"sheets": [{"name": "Juan Perez"}]}, self.outdir)

    def test_no_sheets_writes_nothing(self):
        self.assertEqual(bs.build_from_payload(self.payload(sheets=[]), self.outdir), [])


class CommandLineTests(unittest.TestCase):
    """The two entry points the listener and a human use."""

    def setUp(self):
        self.outdir = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, self.outdir)

    def run_script(self, *args, stdin=None):
        return subprocess.run(
            [sys.executable, SCRIPT, *args],
            input=stdin, capture_output=True, text=True,
        )

    def test_blank_mode_prints_the_path_it_wrote(self):
        r = self.run_script("Juan Perez", "06082026", "--outdir", self.outdir)

        self.assertEqual(r.returncode, 0, r.stderr)
        path = r.stdout.strip()
        self.assertTrue(os.path.exists(path), path)
        self.assertTrue(path.endswith("Juan_Perez_06082026_schedule.pdf"))

    def test_json_on_stdin_prints_one_path_per_sheet(self):
        payload = {
            "pay_period": "2026-06-08",
            "start_date": "2026-06-08",
            "sheets": [
                {"name": "Juan Perez", "days": full_period(), "extra": [], "total": "8:30"},
                {"name": "Maria Lopez", "days": full_period(), "extra": [], "total": ""},
            ],
        }
        r = self.run_script("--json", "-", "--outdir", self.outdir,
                            stdin=json.dumps(payload))

        self.assertEqual(r.returncode, 0, r.stderr)
        paths = r.stdout.split()
        self.assertEqual(len(paths), 2)
        for p in paths:
            self.assertTrue(os.path.exists(p), p)

    def test_json_from_a_file(self):
        path = os.path.join(self.outdir, "payload.json")
        with open(path, "w") as f:
            json.dump({"start_date": "2026-06-08",
                       "sheets": [{"name": "Juan Perez", "days": full_period(),
                                   "extra": [], "total": "8:30"}]}, f)

        r = self.run_script("--json", path, "--outdir", self.outdir)
        self.assertEqual(r.returncode, 0, r.stderr)
        self.assertTrue(os.path.exists(r.stdout.strip()))

    def test_missing_startdate_is_a_usage_error(self):
        r = self.run_script("Juan Perez")
        self.assertNotEqual(r.returncode, 0)
        self.assertIn("startdate", r.stderr)

    def test_empty_payload_reports_no_sheets(self):
        r = self.run_script("--json", "-", "--outdir", self.outdir,
                            stdin='{"start_date": "2026-06-08", "sheets": []}')
        self.assertEqual(r.returncode, 1)
        self.assertIn("No sheets", r.stderr)


if __name__ == "__main__":
    unittest.main()
