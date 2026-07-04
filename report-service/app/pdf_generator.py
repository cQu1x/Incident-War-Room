from io import BytesIO
import re
from xml.sax.saxutils import escape
from datetime import datetime

from reportlab.lib import colors
from reportlab.lib.enums import TA_CENTER, TA_LEFT
from reportlab.lib.pagesizes import A4
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.lib.units import inch
from reportlab.pdfgen import canvas
from reportlab.platypus import (
    HRFlowable,
    KeepTogether,
    Paragraph,
    SimpleDocTemplate,
    Spacer,
    Table,
    TableStyle,
)

from .ai_client import generate_ai_summary, generate_ai_title
from .schemas import ReportRequest


COLOR_TEXT = colors.HexColor("#111827")
COLOR_MUTED = colors.HexColor("#6B7280")
COLOR_LINE = colors.HexColor("#E5E7EB")
COLOR_CARD = colors.HexColor("#F9FAFB")
COLOR_HEADER_BG = colors.HexColor("#F3F4F6")
COLOR_BLUE = colors.HexColor("#2563EB")
COLOR_GREEN = colors.HexColor("#047857")
COLOR_GREEN_BG = colors.HexColor("#D1FAE5")
COLOR_RED = colors.HexColor("#B91C1C")
COLOR_RED_BG = colors.HexColor("#FEE2E2")
COLOR_ORANGE = colors.HexColor("#B45309")
COLOR_ORANGE_BG = colors.HexColor("#FEF3C7")
COLOR_GRAY_BG = colors.HexColor("#E5E7EB")


class NumberedCanvas(canvas.Canvas):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self._saved_page_states = []

    def showPage(self):
        self._saved_page_states.append(dict(self.__dict__))
        self._startPage()

    def save(self):
        page_count = len(self._saved_page_states)

        for state in self._saved_page_states:
            self.__dict__.update(state)
            self.draw_page_number(page_count)
            super().showPage()

        super().save()

    def draw_page_number(self, page_count: int):
        width, _ = A4
        self.saveState()
        self.setFont("Helvetica", 8)
        self.setFillColor(COLOR_MUTED)
        self.drawRightString(
            width - 0.75 * inch,
            0.42 * inch,
            f"Page {self._pageNumber} of {page_count}",
        )
        self.restoreState()


def clean_summary_text(text: str, original_title: str) -> str:
    text = text.replace("■", "-")
    text = text.replace("–", "-")
    text = text.replace("—", "-")

    lines = []

    for line in text.splitlines():
        stripped = line.strip()

        if not stripped:
            lines.append("")
            continue

        if stripped in {"---", "***", "___"}:
            continue

        lower_line = stripped.lower()
        lower_original_title = original_title.lower().strip()

        if "incident summary" in lower_line and lower_original_title in lower_line:
            continue

        lines.append(stripped)

    return "\n".join(lines).strip()


def format_inline_markdown(text: str) -> str:
    safe_text = escape(text)

    safe_text = re.sub(
        r"\*\*(.+?)\*\*",
        r"<b>\1</b>",
        safe_text,
    )

    safe_text = re.sub(
        r"`(.+?)`",
        r'<font name="Courier">\1</font>',
        safe_text,
    )

    return safe_text


def get_badge_colors(value: str) -> tuple:
    normalized_value = value.lower().strip()

    if normalized_value in {"resolved", "closed", "done", "fixed"}:
        return COLOR_GREEN_BG, COLOR_GREEN

    if normalized_value in {"high", "critical", "sev1", "sev-1"}:
        return COLOR_RED_BG, COLOR_RED

    if normalized_value in {"medium", "warning", "sev2", "sev-2"}:
        return COLOR_ORANGE_BG, COLOR_ORANGE

    return COLOR_GRAY_BG, COLOR_TEXT


def make_badge(text: str, styles: dict) -> Table:
    background_color, text_color = get_badge_colors(text)

    badge = Table(
        [[Paragraph(escape(text).upper(), styles["Badge"])]],
        hAlign="LEFT",
    )
    badge.setStyle(
        TableStyle(
            [
                ("BACKGROUND", (0, 0), (-1, -1), background_color),
                ("TEXTCOLOR", (0, 0), (-1, -1), text_color),
                ("BOX", (0, 0), (-1, -1), 0.25, background_color),
                ("LEFTPADDING", (0, 0), (-1, -1), 7),
                ("RIGHTPADDING", (0, 0), (-1, -1), 7),
                ("TOPPADDING", (0, 0), (-1, -1), 3),
                ("BOTTOMPADDING", (0, 0), (-1, -1), 3),
            ]
        )
    )

    return badge


def build_section_title(title: str, styles: dict) -> list:
    return [
        Spacer(1, 12),
        Paragraph(title, styles["Heading"]),
        HRFlowable(
            width="100%",
            thickness=0.7,
            color=COLOR_LINE,
            spaceBefore=0,
            spaceAfter=8,
        ),
    ]

def format_duration(created_at: str, closed_at: str) -> str:
    try:
        start = datetime.fromisoformat(created_at.replace("Z", "+00:00"))
        end = datetime.fromisoformat(closed_at.replace("Z", "+00:00"))
    except ValueError:
        return ""

    total_seconds = max(int((end - start).total_seconds()), 0)
    days, remainder = divmod(total_seconds, 86400)
    hours, remainder = divmod(remainder, 3600)
    minutes, _ = divmod(remainder, 60)

    parts = []
    if days:
        parts.append(f"{days}d")
    if hours:
        parts.append(f"{hours}h")
    if minutes or not parts:
        parts.append(f"{minutes}m")

    return " ".join(parts)


def build_metadata_table(data: ReportRequest, styles: dict) -> Table:
    rows = [
        [Paragraph("Field", styles["TableHeader"]), Paragraph("Value", styles["TableHeader"])],
        [Paragraph("ID", styles["TableLabel"]), Paragraph(escape(data.incident.id), styles["TableValue"])],
        [
            Paragraph("Created at", styles["TableLabel"]),
            Paragraph(escape(data.incident.createdAt), styles["TableValue"]),
        ],
    ]

    if data.incident.closedAt:
        rows.append(
            [
                Paragraph("Closed at", styles["TableLabel"]),
                Paragraph(escape(data.incident.closedAt), styles["TableValue"]),
            ]
        )
        duration_text = format_duration(data.incident.createdAt, data.incident.closedAt)
        if duration_text:
            rows.append(
                [
                    Paragraph("Duration", styles["TableLabel"]),
                    Paragraph(escape(duration_text), styles["TableValue"]),
                ]
            )

    if data.incident.severity:
        rows.append([Paragraph("Severity", styles["TableLabel"]), make_badge(data.incident.severity, styles)])


    table = Table(rows, colWidths=[1.65 * inch, 4.65 * inch], hAlign="LEFT")
    table.setStyle(
        TableStyle(
            [
                ("BACKGROUND", (0, 0), (-1, 0), COLOR_HEADER_BG),
                ("BACKGROUND", (0, 1), (-1, -1), COLOR_CARD),
                ("BOX", (0, 0), (-1, -1), 0.5, COLOR_LINE),
                ("INNERGRID", (0, 0), (-1, -1), 0.35, COLOR_LINE),
                ("VALIGN", (0, 0), (-1, -1), "MIDDLE"),
                ("LEFTPADDING", (0, 0), (-1, -1), 9),
                ("RIGHTPADDING", (0, 0), (-1, -1), 9),
                ("TOPPADDING", (0, 0), (-1, -1), 7),
                ("BOTTOMPADDING", (0, 0), (-1, -1), 7),
            ]
        )
    )

    return table


def build_participants_table(data: ReportRequest, styles: dict) -> Table | Paragraph:
    if not data.participants:
        return Paragraph("No participants provided.", styles["CustomBody"])

    rows = [[Paragraph("Username", styles["TableHeader"]), Paragraph("Telegram ID", styles["TableHeader"])]]

    for participant in data.participants:
        rows.append(
            [
                Paragraph(escape(participant.username), styles["TableValue"]),
                Paragraph(str(participant.userId), styles["TableValue"]),
            ]
        )

    table = Table(rows, colWidths=[3.5 * inch, 2.8 * inch], hAlign="LEFT")
    table.setStyle(
        TableStyle(
            [
                ("BACKGROUND", (0, 0), (-1, 0), COLOR_HEADER_BG),
                ("BACKGROUND", (0, 1), (-1, -1), COLOR_CARD),
                ("BOX", (0, 0), (-1, -1), 0.5, COLOR_LINE),
                ("INNERGRID", (0, 0), (-1, -1), 0.35, COLOR_LINE),
                ("VALIGN", (0, 0), (-1, -1), "MIDDLE"),
                ("LEFTPADDING", (0, 0), (-1, -1), 9),
                ("RIGHTPADDING", (0, 0), (-1, -1), 9),
                ("TOPPADDING", (0, 0), (-1, -1), 7),
                ("BOTTOMPADDING", (0, 0), (-1, -1), 7),
            ]
        )
    )

    return table


def add_summary_to_elements(summary: str, elements: list, styles: dict) -> None:
    current_block = []

    def flush_block() -> None:
        if not current_block:
            return

        card = Table([[current_block.copy()]], colWidths=[6.3 * inch], hAlign="LEFT")
        card.setStyle(
            TableStyle(
                [
                    ("BACKGROUND", (0, 0), (-1, -1), COLOR_CARD),
                    ("BOX", (0, 0), (-1, -1), 0.5, COLOR_LINE),
                    ("LEFTPADDING", (0, 0), (-1, -1), 10),
                    ("RIGHTPADDING", (0, 0), (-1, -1), 10),
                    ("TOPPADDING", (0, 0), (-1, -1), 8),
                    ("BOTTOMPADDING", (0, 0), (-1, -1), 8),
                    ("VALIGN", (0, 0), (-1, -1), "TOP"),
                ]
            )
        )
        elements.append(KeepTogether(card))
        elements.append(Spacer(1, 7))
        current_block.clear()

    for line in summary.splitlines():
        line = line.strip()

        if not line:
            current_block.append(Spacer(1, 4))
            continue

        markdown_heading_match = re.match(r"^#{1,6}\s*(.+)$", line)
        numbered_heading_match = re.match(r"^\*\*(\d+\.\s+.+?)\*\*$", line)
        plain_numbered_heading_match = re.match(r"^(\d+\.\s+.+)$", line)
        bold_heading_match = re.match(r"^\*\*(.+?)\*\*$", line)

        if markdown_heading_match or numbered_heading_match or bold_heading_match or plain_numbered_heading_match:
            flush_block()

            heading_text = (
                markdown_heading_match.group(1)
                if markdown_heading_match
                else numbered_heading_match.group(1)
                if numbered_heading_match
                else bold_heading_match.group(1)
                if bold_heading_match
                else plain_numbered_heading_match.group(1)
            )
            current_block.append(Paragraph(format_inline_markdown(heading_text), styles["CustomSectionHeading"]))
            continue

        if line.startswith("- ") or line.startswith("• "):
            bullet_text = line[2:].strip()
            current_block.append(
                Paragraph(
                    format_inline_markdown(bullet_text),
                    styles["CustomBullet"],
                    bulletText="•",
                )
            )
            continue

        current_block.append(
            Paragraph(
                format_inline_markdown(line),
                styles["CustomBody"],
            )
        )

    flush_block()


def make_header_footer(title: str):
    def draw(canvas_obj, doc):
        width, height = A4
        canvas_obj.saveState()

        canvas_obj.setFont("Helvetica-Bold", 8)
        canvas_obj.setFillColor(COLOR_MUTED)
        canvas_obj.drawString(0.75 * inch, height - 0.42 * inch, "Incident War Room Report")


        canvas_obj.setStrokeColor(COLOR_LINE)
        canvas_obj.setLineWidth(0.4)
        canvas_obj.line(0.75 * inch, height - 0.52 * inch, width - 0.75 * inch, height - 0.52 * inch)


        canvas_obj.restoreState()

    return draw


def generate_incident_report(data: ReportRequest) -> bytes:
    buffer = BytesIO()

    document = SimpleDocTemplate(
        buffer,
        pagesize=A4,
        rightMargin=0.75 * inch,
        leftMargin=0.75 * inch,
        topMargin=0.8 * inch,
        bottomMargin=0.75 * inch,
    )

    base_styles = getSampleStyleSheet()

    styles = {
        "Title": ParagraphStyle(
            name="CustomTitle",
            parent=base_styles["Title"],
            fontName="Helvetica-Bold",
            fontSize=23,
            leading=29,
            alignment=TA_CENTER,
            spaceAfter=6,
            textColor=COLOR_TEXT,
        ),
        "Subtitle": ParagraphStyle(
            name="CustomSubtitle",
            parent=base_styles["Normal"],
            fontName="Helvetica",
            fontSize=9.5,
            leading=13,
            alignment=TA_CENTER,
            spaceAfter=18,
            textColor=COLOR_MUTED,
        ),
        "Heading": ParagraphStyle(
            name="CustomHeading",
            parent=base_styles["Heading2"],
            fontName="Helvetica-Bold",
            fontSize=15.5,
            leading=19,
            spaceBefore=4,
            spaceAfter=3,
            textColor=COLOR_TEXT,
        ),
        "CustomSectionHeading": ParagraphStyle(
            name="CustomSectionHeading",
            parent=base_styles["Heading3"],
            fontName="Helvetica-Bold",
            fontSize=12.8,
            leading=16,
            spaceBefore=0,
            spaceAfter=5,
            textColor=COLOR_BLUE,
        ),
        "CustomBody": ParagraphStyle(
            name="CustomBody",
            parent=base_styles["Normal"],
            fontName="Helvetica",
            fontSize=11,
            leading=15.5,
            alignment=TA_LEFT,
            spaceAfter=4,
            textColor=COLOR_TEXT,
        ),
        "CustomBullet": ParagraphStyle(
            name="CustomBullet",
            parent=base_styles["Normal"],
            fontName="Helvetica",
            fontSize=11,
            leading=15.5,
            leftIndent=14,
            firstLineIndent=0,
            bulletIndent=0,
            spaceAfter=3,
            textColor=COLOR_TEXT,
        ),
        "TableHeader": ParagraphStyle(
            name="TableHeader",
            parent=base_styles["Normal"],
            fontName="Helvetica-Bold",
            fontSize=10,
            leading=13,
            textColor=COLOR_TEXT,
        ),
        "TableLabel": ParagraphStyle(
            name="TableLabel",
            parent=base_styles["Normal"],
            fontName="Helvetica-Bold",
            fontSize=10.5,
            leading=14,
            textColor=COLOR_MUTED,
        ),
        "TableValue": ParagraphStyle(
            name="TableValue",
            parent=base_styles["Normal"],
            fontName="Helvetica",
            fontSize=10.5,
            leading=14,
            textColor=COLOR_TEXT,
        ),
        "Badge": ParagraphStyle(
            name="Badge",
            parent=base_styles["Normal"],
            fontName="Helvetica-Bold",
            fontSize=8.5,
            leading=10,
            alignment=TA_CENTER,
            textColor=COLOR_TEXT,
        ),
    }

    elements = []

    ai_title = generate_ai_title(data)
    safe_ai_title = escape(ai_title)
    elements.append(Paragraph(safe_ai_title, styles["Title"]))

    elements.extend(build_section_title("Incident Metadata", styles))
    elements.append(build_metadata_table(data, styles))

    elements.extend(build_section_title("Participants", styles))
    elements.append(build_participants_table(data, styles))

    ai_summary = generate_ai_summary(data)
    ai_summary = clean_summary_text(
        text=ai_summary,
        original_title=data.incident.title,
    )

    elements.extend(build_section_title("AI Summary", styles))
    add_summary_to_elements(ai_summary, elements, styles)

    document.build(
        elements,
        onFirstPage=make_header_footer(ai_title),
        onLaterPages=make_header_footer(ai_title),
        canvasmaker=NumberedCanvas,
    )

    pdf_bytes = buffer.getvalue()
    buffer.close()

    return pdf_bytes
