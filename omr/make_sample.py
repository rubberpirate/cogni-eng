"""Generate an original CC0 filled test sheet and matching native OMRChecker template."""
import json
from pathlib import Path
from PIL import Image, ImageDraw, ImageFont


def generate(destination: Path) -> None:
    destination.mkdir(parents=True, exist_ok=True)
    image = Image.new("RGB", (1200, 1600), "white")
    draw = ImageDraw.Draw(image)
    font = ImageFont.load_default(size=23)
    small = ImageFont.load_default(size=18)
    draw.text((80, 55), "OMRChecker - filled practice sheet", fill="#263846", font=ImageFont.load_default(size=35))
    draw.text((80, 110), "Original CC0 sample | synthetic learner | no real student data", fill="#526575", font=small)
    draw.text((80, 170), "Student name: Synthetic Test Learner", fill="#263846", font=font)
    draw.text((80, 215), "Roll: 123456", fill="#263846", font=font)
    blocks = {}
    expected = {"Roll": "123456"}

    def bubble(x: int, y: int, text: str, filled: bool) -> None:
        draw.ellipse((x-13, y-13, x+13, y+13), fill="#252525" if filled else "white", outline="#606060", width=2)
        if not filled:
            draw.text((x, y), text, anchor="mm", fill="#686868", font=small)

    for column, digit in enumerate("123456"):
        x = 365 + column * 65
        for value in range(10):
            bubble(x, 270+value*35, str(value), str(value)==digit)
        blocks[f"roll_{column+1}"] = {"fieldType":"QTYPE_INT", "origin":[x-11,259], "bubbleDimensions":[22,22], "bubblesGap":35, "labelsGap":0,"fieldLabels":[f"roll{column+1}"]}
    draw.text((80, 650), "Marked responses (A-D). Question 7 is blank; question 12 has two marks.", fill="#263846", font=small)
    answers = ["A","C","D","B","A","D","","B","D","A","B","AC","A","D","C","B","A","C","D","B"]
    for index, answer in enumerate(answers):
        column, row = divmod(index, 10)
        x, y = 180+column*560, 750+row*70
        draw.text((x-72,y),f"{index+1:02d}",anchor="mm",fill="#263846",font=font)
        for option, label in enumerate("ABCD"):
            bubble(x+option*70,y,label,label in answer)
        blocks[f"answer_{index+1}"]={"fieldType":"QTYPE_MCQ4","origin":[x-11,y-11],"bubbleDimensions":[22,22],"bubblesGap":70,"labelsGap":0,"fieldLabels":[f"q{index+1}"]}
        expected[f"q{index+1}"]=answer
    draw.text((80,1515),"Use the accompanying template.json. Expected answers are in expected-answers.json.",fill="#526575",font=small)
    template={"pageDimensions":[1200,1600],"bubbleDimensions":[22,22],"preProcessors":[],"customLabels":{"Roll":["roll1..6"]},"fieldBlocks":blocks}
    image.save(destination / "filled-sheet.png", optimize=True)
    (destination/"template.json").write_text(json.dumps(template,indent=2)+"\n")
    (destination/"expected-answers.json").write_text(json.dumps({"notice":"Synthetic test only. These are filled responses, not a grading key.","responses":expected},indent=2)+"\n")


if __name__ == "__main__":
    generate(Path(__file__).resolve().parents[1]/"apps/web/public/samples/omrchecker")
