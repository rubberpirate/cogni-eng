"""Protocol guard tests; these do not prove successful GPU inference."""
import unittest
import base64
import io
from unittest.mock import patch

import server
from PIL import Image, ImageDraw, ImageFont


class RequestBounds(unittest.TestCase):
    def test_untrusted_locations_and_instruction_mode_are_rejected(self):
        image=Image.new("RGB",(100,100),"white")
        data=io.BytesIO();image.save(data,format="PNG")
        uri="data:image/png;base64,"+base64.b64encode(data.getvalue()).decode()
        cases = [
            {"model":"lightonai/LightOnOCR-2-1B","messages":[{"role":"user","content":[{"type":"image_url","image_url":{"url":uri}},{"type":"text","text":"Untrusted instruction"}]}]},
            {"model": "untrusted/model", "messages": []},
            {"model": next(iter(server.MODELS)), "messages": []},
            {"model": "lightonai/LightOnOCR-2-1B", "messages": [{"role": "user", "content": [{"type": "image_url", "image_url": {"url": "https://untrusted.invalid/image.png"}}]}]},
        ]
        for payload in cases:
            with self.subTest(payload=payload), patch.object(server, "load_model") as load:
                with self.assertRaises(ValueError):
                    server.infer(payload)
                load.assert_not_called()


class BubbleArtifact(unittest.TestCase):
    def test_original_classifier_separates_printed_and_filled_marks(self):
        image=Image.new("RGB",(500,100),"white");draw=ImageDraw.Draw(image)
        draw.ellipse((30,30,70,70),outline=(190,120,120),width=2)
        draw.text((50,50),"3",font=ImageFont.load_default(size=24),fill=(190,120,120),anchor="mm")
        draw.ellipse((130,30,170,70),outline=(190,120,120),width=2)
        draw.ellipse((134,34,166,66),fill="black")
        from bubbles import classify
        result=classify(image,[{"x":.1,"y":.5},{"x":.3,"y":.5}],.044)
        self.assertEqual(result["classes"],[0,1])


if __name__ == "__main__":
    unittest.main()
