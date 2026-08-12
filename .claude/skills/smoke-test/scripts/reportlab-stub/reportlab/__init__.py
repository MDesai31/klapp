"""Stand-in for reportlab, so build_schedule.py can be exercised on a box
that has no reportlab (and no pip to install one).

Put this directory on PYTHONPATH and build_schedule.py runs end to end, but
instead of drawing a PDF it appends the table it *would* have drawn to
$SCHEDULE_STUB_DUMP as plain text. The .pdf files it names are still created
(one line of placeholder text), so anything downstream that only checks paths
is happy.

See the "Test the print path" section of the smoke-test skill.
"""

__version__ = "stub"
Version = "stub"
