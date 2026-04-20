"""PDF export implementation using reportlab.

Features:
- Optional title rendered at the top of the first page
- Table with styled header row
- Alternating row colours for readability
- Auto-scaled column widths
- Page numbers in the footer
"""

import io
from typing import Any, List, Optional

from reportlab.lib import colors
from reportlab.lib.pagesizes import A4, landscape
from reportlab.lib.styles import getSampleStyleSheet
from reportlab.lib.units import cm
from reportlab.platypus import (
    Paragraph,
    SimpleDocTemplate,
    Spacer,
    Table,
    TableStyle,
)


_HEADER_BG = colors.HexColor("#2F5496")
_HEADER_FG = colors.white
_ROW_EVEN_BG = colors.HexColor("#EEF2FF")
_ROW_ODD_BG = colors.white
_GRID_COLOR = colors.HexColor("#CCCCCC")


def _page_number_footer(canvas, doc):
    """Draw a page number at the bottom centre of each page."""
    canvas.saveState()
    canvas.setFont("Helvetica", 8)
    page_text = f"Page {doc.page}"
    canvas.drawCentredString(doc.pagesize[0] / 2.0, 0.75 * cm, page_text)
    canvas.restoreState()


class PdfExporter:
    def export(
        self,
        headers: List[str],
        rows: List[List[Any]],
        title: Optional[str] = None,
    ) -> bytes:
        """Serialise *headers* and *rows* to a PDF document in memory."""
        buf = io.BytesIO()
        page_size = landscape(A4) if len(headers) > 6 else A4

        doc = SimpleDocTemplate(
            buf,
            pagesize=page_size,
            rightMargin=1.5 * cm,
            leftMargin=1.5 * cm,
            topMargin=2 * cm,
            bottomMargin=2 * cm,
        )

        styles = getSampleStyleSheet()
        story = []

        # Optional title
        if title:
            title_style = styles["Title"]
            story.append(Paragraph(title, title_style))
            story.append(Spacer(1, 0.4 * cm))

        # Build table data: header row + data rows
        table_data: List[List[str]] = [
            [str(h) for h in headers]
        ]
        for row in rows:
            table_data.append([
                "" if cell is None else str(cell)
                for cell in row
            ])

        # Distribute available width evenly across columns
        usable_width = page_size[0] - doc.leftMargin - doc.rightMargin
        num_cols = max(len(headers), 1)
        col_width = usable_width / num_cols

        table = Table(table_data, colWidths=[col_width] * num_cols, repeatRows=1)

        # Build table style
        table_style_cmds = [
            # Header row
            ("BACKGROUND", (0, 0), (-1, 0), _HEADER_BG),
            ("TEXTCOLOR", (0, 0), (-1, 0), _HEADER_FG),
            ("FONTNAME", (0, 0), (-1, 0), "Helvetica-Bold"),
            ("FONTSIZE", (0, 0), (-1, 0), 9),
            ("ALIGN", (0, 0), (-1, 0), "CENTER"),
            ("VALIGN", (0, 0), (-1, 0), "MIDDLE"),
            ("BOTTOMPADDING", (0, 0), (-1, 0), 6),
            ("TOPPADDING", (0, 0), (-1, 0), 6),
            # Data rows
            ("FONTNAME", (0, 1), (-1, -1), "Helvetica"),
            ("FONTSIZE", (0, 1), (-1, -1), 8),
            ("VALIGN", (0, 1), (-1, -1), "TOP"),
            ("TOPPADDING", (0, 1), (-1, -1), 4),
            ("BOTTOMPADDING", (0, 1), (-1, -1), 4),
            # Grid
            ("GRID", (0, 0), (-1, -1), 0.4, _GRID_COLOR),
            ("BOX", (0, 0), (-1, -1), 0.6, _GRID_COLOR),
        ]

        # Alternating row background colours
        for row_idx in range(1, len(table_data)):
            bg = _ROW_EVEN_BG if row_idx % 2 == 0 else _ROW_ODD_BG
            table_style_cmds.append(
                ("BACKGROUND", (0, row_idx), (-1, row_idx), bg)
            )

        table.setStyle(TableStyle(table_style_cmds))
        story.append(table)

        doc.build(story, onFirstPage=_page_number_footer, onLaterPages=_page_number_footer)
        return buf.getvalue()
