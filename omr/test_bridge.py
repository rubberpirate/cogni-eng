import json
from pathlib import Path
import unittest
from unittest.mock import patch
import subprocess
import tempfile

from bridge import InputError, bounded_template, decode_values, read_sheet, read_identity


def minimal():
    return {"pageDimensions":[100,100],"bubbleDimensions":[10,10],"fieldBlocks":{"answers":{"fieldLabels":["q1..2"],"fieldType":"QTYPE_MCQ4","origin":[10,10],"bubblesGap":15,"labelsGap":20}}}


class TemplateBoundaryTests(unittest.TestCase):
    def test_missing_or_invalid_engine_revision_fails_before_import(self):
        sample = Path(__file__).resolve().parents[1] / "apps/web/public/samples/omrchecker/filled-sheet.png"
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "src").mkdir()
            (root / "src/core.py").write_text("raise AssertionError('must not import')")
            for revision in (None, "not-an-upstream-revision"):
                with self.subTest(revision=revision):
                    if revision is not None:
                        (root / "REVISION").write_text(revision)
                    with self.assertRaisesRegex(InputError, "revision"):
                        read_sheet(root, sample, {"template": minimal()})

    def test_identity_boxes_are_bounded_and_ocr_failure_keeps_manual_review(self):
        import numpy as np
        image = np.full((100, 100), 255, dtype=np.uint8)
        region = {"kind": "student_name", "rect": {"x": .1, "y": .1, "width": .5, "height": .2}}
        with tempfile.TemporaryDirectory() as directory:
            work = Path(directory)
            for invalid in [[region, region], [{**region, "rect": {**region["rect"], "width": 2}}], [None]]:
                with self.assertRaises(InputError): read_identity(image, invalid, work)
            with patch("bridge.subprocess.run", side_effect=FileNotFoundError):
                result = read_identity(image, [region], work)[0]
            self.assertTrue(result["uncertain"])
            self.assertEqual(result["text"], "")
            self.assertIn("unavailable", result["reason"])
            tsv = "header\n" + "\t".join(["5", "1", "1", "1", "1", "1", "0", "0", "40", "10", "42", "Name: Learner"])
            with patch("bridge.subprocess.run", return_value=subprocess.CompletedProcess([], 0, stdout=tsv)):
                result = read_identity(image, [region], work)[0]
            self.assertEqual(result["text"], "Learner")
            self.assertTrue(result["uncertain"])
            self.assertIn("low", result["reason"])

    def test_resource_paths_are_replaced_with_uploaded_resources(self):
        template=minimal()
        template["preProcessors"]=[{"name":"CropOnMarkers","options":{"relativePath":"../../private-file"}},{"name":"FeatureBasedAlignment","options":{"reference":"/private-file"}}]
        safe=bounded_template(template,True,True)
        self.assertEqual(safe["preProcessors"][0]["options"]["relativePath"],"marker.png")
        self.assertEqual(safe["preProcessors"][1]["options"]["reference"],"reference.png")
        self.assertEqual(template["preProcessors"][0]["options"]["relativePath"],"../../private-file")
        with self.assertRaises(InputError): bounded_template(template,False,False)
        template=minimal();template['customLabels']={'Roll':['roll1..999999999999']}
        with self.assertRaises(InputError): bounded_template(template,False,False)
        template=minimal();template['outputColumns']=['q1..999999999999']
        with self.assertRaises(InputError): bounded_template(template,False,False)

    def test_oversized_ranges_and_kernels_are_rejected_before_upstream_import(self):
        for label in ["q1..999999999999", "q8..2"]:
            with self.subTest(label=label):
                template=minimal();template["fieldBlocks"]["answers"]["fieldLabels"]=[label]
                with self.assertRaises(InputError): bounded_template(template,False,False)
        template=minimal();template["preProcessors"]=[{"name":"CropPage","options":{"morphKernel":[1000000,1000000]}}]
        with self.assertRaises(InputError): bounded_template(template,False,False)
        template["preProcessors"]=[{"name":"ArbitraryPlugin","options":{}}]
        with self.assertRaises(InputError): bounded_template(template,False,False)

    def test_multiple_values_and_ambiguous_concatenation(self):
        self.assertEqual(decode_values("AC",list("ABCD")),["A","C"])
        self.assertEqual(decode_values("",list("ABCD")),[])
        self.assertIsNone(decode_values("AB",["A","AB","B"]))
        self.assertIsNone(decode_values("Z",list("ABCD")))

    def test_bundled_engine_reads_all_original_sample_fields_without_git(self):
        root = Path(__file__).resolve().parent / "engine"
        sample=Path(__file__).resolve().parents[1]/"apps/web/public/samples/omrchecker"
        source=(sample/"template.json").read_bytes()
        with patch("bridge.subprocess.run", wraps=subprocess.run) as commands:
            result=read_sheet(root,sample/"filled-sheet.png",{"template":json.loads(source)})
        self.assertFalse(any(call.args[0][0] == "git" for call in commands.call_args_list))
        expected=json.loads((sample/"expected-answers.json").read_text())["responses"]
        self.assertEqual(result["responses"],expected)
        self.assertEqual(result["engine"],"OMRChecker")
        self.assertEqual(result["revision"], "5cf44a5a9c7a49e2e541c6e19aefd09a41a4d494")
        self.assertEqual(len(result["fields"]),26)
        self.assertEqual([f["label"] for f in result["fields"] if f["uncertain"]],["q7","q12"])
        self.assertTrue(result["multiple_marked"])
        self.assertEqual((sample/"template.json").read_bytes(),source)


if __name__=="__main__": unittest.main()
