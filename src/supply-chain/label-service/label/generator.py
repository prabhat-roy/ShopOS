"""Label generation logic for ZPL and human-readable text formats."""

import base64
import io
from datetime import datetime, timezone

import barcode
from barcode.writer import SVGWriter

from label.models import LabelRequest, LabelResponse


class LabelGenerator:
    """Generates shipping labels in ZPL II and plain-text formats."""

    def generate_zpl(self, request: LabelRequest) -> str:
        """
        Generate a ZPL II label string suitable for thermal printer output.

        Uses standard ZPL II commands:
          ^XA / ^XZ  — label start/end
          ^FO        — field origin (x, y)
          ^FD        — field data
          ^A0N       — scalable font
          ^BY        — bar code field default
          ^BC        — Code 128 barcode
          ^GB        — graphic box (separator lines)
          ^FS        — field separator
        """
        tn = request.tracking_number
        carrier = request.carrier.upper()
        from_addr = request.from_address
        to_addr = request.to_address
        weight_str = f"{request.weight:.2f} LB"

        # Dimensions line (optional)
        dims_str = ""
        if request.dimensions:
            l = request.dimensions.get("length", "")
            w = request.dimensions.get("width", "")
            h = request.dimensions.get("height", "")
            if l and w and h:
                dims_str = f"^FO50,610^A0N,22,22^FDDims: {l}x{w}x{h} in^FS\n"

        from_street2 = f"\n^FO50,195^A0N,22,22^FD{from_addr.street2}^FS" if from_addr.street2 else ""
        to_street2 = f"\n^FO50,430^A0N,26,26^FD{to_addr.street2}^FS" if to_addr.street2 else ""

        zpl = (
            "^XA\n"
            # Border
            "^GB800,1200,4^FS\n"
            # Carrier name header
            f"^FO50,30^A0N,55,55^FD{carrier}^FS\n"
            "^FO50,95^GB700,3,3^FS\n"
            # FROM section
            "^FO50,110^A0N,28,28^FDFROM:^FS\n"
            f"^FO50,145^A0N,24,24^FD{from_addr.name}^FS\n"
            f"^FO50,170^A0N,22,22^FD{from_addr.street1}^FS\n"
            f"{from_street2}"
            f"^FO50,220^A0N,22,22^FD{from_addr.city}, {from_addr.state} {from_addr.postal_code}^FS\n"
            f"^FO50,245^A0N,22,22^FD{from_addr.country}^FS\n"
            "^FO50,275^GB700,3,3^FS\n"
            # TO section
            "^FO50,295^A0N,32,32^FDSHIP TO:^FS\n"
            f"^FO50,340^A0N,30,30^FD{to_addr.name}^FS\n"
            f"^FO50,380^A0N,28,28^FD{to_addr.street1}^FS\n"
            f"{to_street2}"
            f"^FO50,460^A0N,28,28^FD{to_addr.city}, {to_addr.state} {to_addr.postal_code}^FS\n"
            f"^FO50,495^A0N,26,26^FD{to_addr.country}^FS\n"
            "^FO50,530^GB700,3,3^FS\n"
            # Weight
            f"^FO50,550^A0N,26,26^FDWeight: {weight_str}^FS\n"
            f"{dims_str}"
            # Shipment ID
            f"^FO50,640^A0N,22,22^FDShipment ID: {request.shipment_id}^FS\n"
            "^FO50,670^GB700,3,3^FS\n"
            # Barcode (Code 128)
            f"^FO50,690^BY2,3,80^BCN,80,Y,N,N^FD{tn}^FS\n"
            "^FO50,790^GB700,3,3^FS\n"
            # Human-readable tracking number below barcode
            f"^FO50,810^A0N,36,36^FD{tn}^FS\n"
            "^XZ"
        )
        return zpl

    def generate_text(self, request: LabelRequest) -> str:
        """
        Generate a human-readable ASCII shipping label.
        """
        from_addr = request.from_address
        to_addr = request.to_address

        from_street2_line = f"  {from_addr.street2}\n" if from_addr.street2 else ""
        to_street2_line = f"  {to_addr.street2}\n" if to_addr.street2 else ""
        dims_line = ""
        if request.dimensions:
            l = request.dimensions.get("length", "")
            w = request.dimensions.get("width", "")
            h = request.dimensions.get("height", "")
            if l and w and h:
                dims_line = f"  Dimensions : {l} x {w} x {h} in\n"

        label = (
            "=" * 60 + "\n"
            f"  CARRIER: {request.carrier.upper()}\n"
            "=" * 60 + "\n"
            "\n"
            "  FROM:\n"
            f"  {from_addr.name}\n"
            f"  {from_addr.street1}\n"
            f"{from_street2_line}"
            f"  {from_addr.city}, {from_addr.state} {from_addr.postal_code}\n"
            f"  {from_addr.country}\n"
            "\n"
            "-" * 60 + "\n"
            "\n"
            "  SHIP TO:\n"
            f"  {to_addr.name}\n"
            f"  {to_addr.street1}\n"
            f"{to_street2_line}"
            f"  {to_addr.city}, {to_addr.state} {to_addr.postal_code}\n"
            f"  {to_addr.country}\n"
            "\n"
            "-" * 60 + "\n"
            "\n"
            f"  Weight     : {request.weight:.2f} LB\n"
            f"{dims_line}"
            f"  Shipment ID: {request.shipment_id}\n"
            "\n"
            "-" * 60 + "\n"
            f"  TRACKING: {request.tracking_number}\n"
            "=" * 60 + "\n"
        )
        return label

    def _generate_barcode_base64(self, tracking_number: str) -> str:
        """
        Render a Code 128 barcode as an SVG and return it base64-encoded.
        Falls back to base64-encoding the raw tracking number bytes if
        the python-barcode library is unavailable or fails.
        """
        try:
            code128_class = barcode.get_barcode_class("code128")
            buffer = io.BytesIO()
            code = code128_class(tracking_number, writer=SVGWriter())
            code.write(buffer)
            svg_bytes = buffer.getvalue()
            return base64.b64encode(svg_bytes).decode("utf-8")
        except Exception:
            # Graceful fallback: encode the tracking number directly
            return base64.b64encode(tracking_number.encode("utf-8")).decode("utf-8")

    def generate(self, request: LabelRequest) -> LabelResponse:
        """
        Generate a label for the given request.
        Selects ZPL or TEXT format based on request.label_format.
        """
        fmt = (request.label_format or "ZPL").upper()

        if fmt == "ZPL":
            label_data = self.generate_zpl(request)
        else:
            label_data = self.generate_text(request)

        barcode_data = self._generate_barcode_base64(request.tracking_number)

        return LabelResponse(
            shipmentId=request.shipment_id,
            trackingNumber=request.tracking_number,
            labelFormat=fmt,
            labelData=label_data,
            barcodeData=barcode_data,
            generatedAt=datetime.now(tz=timezone.utc),
        )
