import os

# Where the rendered tables are appended. Deleted between runs by whoever
# is testing, not by this module — the listener builds one sheet per worker
# and they all belong in one dump.
DUMP = os.environ.get("SCHEDULE_STUB_DUMP", "/tmp/klapp-scratch/schedule-rows.txt")


class Paragraph:
    def __init__(self, *args, **kwargs):
        pass


class Frame:
    def __init__(self, *args, **kwargs):
        pass


class PageTemplate:
    def __init__(self, *args, **kwargs):
        pass


class TableStyle:
    def __init__(self, cmds):
        self.cmds = cmds


class Table:
    def __init__(self, data, colWidths=None, rowHeights=None):
        self.data = data
        self.colWidths = colWidths
        self.rowHeights = rowHeights
        self.style = None

    def setStyle(self, style):
        self.style = style


class BaseDocTemplate:
    def __init__(self, filename, **kwargs):
        self.filename = filename

    def addPageTemplates(self, templates):
        pass

    def build(self, flowables):
        tbl = flowables[0]

        os.makedirs(os.path.dirname(DUMP) or ".", exist_ok=True)
        with open(DUMP, "a") as f:
            f.write(f"=== {self.filename}\n")
            # Row and height counts must stay in step, and the row height is
            # what tells you the sheet still fits its half page after extra
            # spill rows were added.
            f.write(f"    rows={len(tbl.data)} heights={len(tbl.rowHeights)} "
                    f"min_row_h={min(tbl.rowHeights):.2f}\n")
            for row in tbl.data:
                cells = [str(c) for c in row]
                f.write("    " + " | ".join(cells).rstrip(" |") + "\n")
            f.write("\n")

        with open(self.filename, "w") as f:
            f.write("stub pdf\n")
