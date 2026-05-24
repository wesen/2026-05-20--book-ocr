---
docType: reference
status: active
intent: short-term
topics:
  - ocr
  - experiments
created: 2026-05-24
updated: 2026-05-24
---

<!-- page:001 -->

Technical Report 794

Presentation Based User Interfaces

Eugene C. Ciccarelli IV

MIT Artificial Intelligence Laboratory


<!-- page:002 -->

This blank page was inserted to preserve pagination.


<!-- page:003 -->

PRESENTATION BASED USER INTERFACES

by

Eugene Charles Ciccarelli IV

B.S., Massachusetts Institute of Technology
(1975)

M.S., Massachusetts Institute of Technology
(1978)

Artificial Intelligence Laboratory
Massachusetts Institute of Technology

August 1984

(C) Massachusetts Institute of Technology 1984

This is a revised version of a thesis submitted to the Department of Electrical Engineering and Computer Science on August 27, 1984, in partial fulfilment of the requirements for the degree of Doctor of Philosophy.

This report describes research done at the Artificial Intelligence Laboratory of the Massachusetts Institute of Technology. Support for the laboratory's artificial intelligence research is provided in part by the Office of Naval Research under Office of Naval Research contract N00014-75-C-0522, in part by the System Development Foundation, and in part by Wang Laboratories.


<!-- page:004 -->

PRESENTATION BASED USER INTERFACES

by

Eugene Charles Ciccarelli IV

Abstract

A prototype presentation system base is described. It offers mechanisms, tools, and ready-made parts for building user interfaces. A general user interface model underlies the base, organized around the concept of a presentation: a visible text or graphic form conveying information. The base and model emphasize domain independence and style independence, to apply to the widest possible range of interfaces.

The primitive presentation system model treats the interface as a system of processes maintaining a semantic relation between an application data base and a presentation data base, the symbolic screen description containing presentations. A presenter continually updates the presentation data base from the application data base. The user manipulates presentations with a presentation editor. A recognizer translates the user's presentation manipulation into application data base commands. The primitive presentation system can be extended to model more complex systems by attaching additional presentation systems. In order to illustrate the model's generality and descriptive capabilities, extended model structures for several existing user interfaces are discussed.

The base provides support for building the application and presentation data bases, linked together into a single, uniform network, including descriptions of classes of objects as well as the objects themselves. The base provides an initial presentation data base network, graphics to continuously display it, and editing functions. A variety of tools and mechanisms help create and control presenters and recognizers. To demonstrate the base's utility, three interfaces to an operating system were constructed, embodying different styles: icon, menu, and graphical annotation.

Thesis Supervisor: Professor Carl Hewitt
Title: Associate Professor of Electrical Engineering and Computer Science

Thesis Supervisor: Dr. Richard Waters
Title: Principal Research Scientist, Artificial Intelligence Laboratory


<!-- page:005 -->

Acknowledgments

My thesis committee, Carl Hewitt, Dick Waters, and Hal Abelson, have been helpful and encouraging. They have all aided significantly in shaping this thesis and improving its quality.

Norton Greenfeld and Martin Yonke introduced me to the world of the presentation concept. It was while working in their group at BBN that I began to think that the concept could serve to explain what was going on in various user interfaces.

Several people have helped with discussions and suggestions at various stages in the development of the ideas, including Lee Blaine, Ron Brachman, Charles Davis, Jeff Gibbons, Earl Killian, Henry Lieberman, Fanya Montalvo, Chuck Rich, Jan Walker, Bill Woods, and Frank Zdybel.

Dan Halbert and Bruce Roberts provided information and the sample screen images for the Xerox Star and Steamer systems, respectively.


<!-- page:006 -->

Table of Contents

Chapter One: Introduction and Overview 8
1.1 The Primitive Presentation System Model 9
1.2 Constructing Larger Presentation System Models 16
1.3 Describing Presentation Systems 17
1.4 PPSbase: A Presentation System Base 18
1.5 Constructing User Interfaces 20
1.6 Related Work 21

Chapter Two: The Primitive Presentation System (PPS) Model 28
2.1 PPSCalc 28
2.2 The Application Data Base 32
2.3 The Presentation Data Base 35
2.4 The Presentation Editor 39
2.5 The Presenter 39
2.6 The Recognizer 43
2.7 The Representation Shift Model and Direct Manipulation 48

Chapter Three: Constructing Larger Presentation System Models 54
3.1 Adding a Planned Data Base 54
3.2 Adding a Data Base of Commands 58
3.3 Adding Interfaces to PPS Components 60
3.4 Shared Screen Space and Presentation Structure 62
3.5 Concluding Remarks 66

Chapter Four: Describing Presentation Systems 67
4.1 Emacs Dired 68
4.2 Zmacs 74
4.3 Xerox Star 80
4.4 Steamer 90
4.5 Summary of Structural Features 97

Chapter Five: PSBase: A Presentation System Base 100
5.1 Data Base Mechanisms 103
5.2 Graphics Redisplay 114
5.3 Presentation Editor Functions 115
5.4 Presenter Support 115
5.5 Recognizer Support 124


<!-- page:007 -->

5.6 Basic Style Packages 127
5.7 Summary 141

Chapter Six: Constructing Presentation Systems 142
6.1 The User's View of the Three Interfaces 142
6.2 Common Implementation Details 167
6.3 Icon-Style Interface Implementation 173
6.4 Menu-Style Interface Implementation 178
6.5 Annotation-Style Interface Implementation 181
6.6 Other Style Possibilities 183
6.7 Summary 184

Chapter Seven: Areas for Further Research 187
7.1 PSBase Limitations 187


<!-- page:008 -->

Table of Figures

Figure 1-1: A Rudimentary User Interface 11
Figure 1-2: The Representation Shift Model 13
Figure 1-3: The Primitive Presentation System (PPS) Model 15
Figure 1-4: Structure of PSBase 19
Figure 2-1: The Primitive Presentation System (PPS) Model 29
Figure 2-2: PPSCalc -- Formula Display 30
Figure 2-3: PPSCalc -- Value Display 30
Figure 2-4: PPSCalc -- After Editing 31
Figure 2-5: PPSCalc -- After Recalculation 31
Figure 2-6: PPSCalc -- New Formulas 31
Figure 2-7: PPSCalc -- Values of New Formulas 32
Figure 2-8: World Model 34
Figure 2-9: Presenter Parts 40
Figure 2-10: Recognizer Parts 44
Figure 2-11: PPSCalc -- Value Moved 45
Figure 2-12: PPSCalc -- Formula Moved 46
Figure 2-13: PPSCalc -- Preparing to Copy Formula 46
Figure 2-14: Representation Shift Model 49
Figure 2-15: Functional Mapping in the PPS Model 52
Figure 3-1: Planned Data Base Extension 56
Figure 3-2: Extension with Both Planning and Immediate Changes 57
Figure 3-3: Command Data Base Extension 59
Figure 3-4: Presenter Interface Extension 61
Figure 3-5: Presenter Commands Extension 63
Figure 4-1: Dired Model 75
Figure 4-2: Zmacs Model 75
Figure 4-3: Zmacs Scroll Bar 81
Figure 4-4: Xerox Star -- Desktop Display 83
Figure 4-5: Xerox Star -- Opened Folder 84
Figure 4-6: Xerox Star -- Property Sheet 86
Figure 4-7: Xerox Star -- Delete Confirmation 87
Figure 4-8: Xerox Star Model 88
Figure 4-9: Sample Streamer Schematic 91
Figure 4-10: Streamer Menu Console 93
Figure 4-11: Steamer Model 94
Figure 4-12: Sample of Steamer Icons 95
Figure 5-1: PPSBase Support of PPS Components 101
Figure 5-2: Structure of PSBase 102
Figure 5-3: A Class Description Network 105


<!-- page:009 -->

Figure 5-4: Sample Presentation Data Base Structure 107
Figure 5-5: Inter-Presentation Relationships 108
Figure 5-6: Command Description Support 110
Figure 5-7: Reference Resolution 113
Figure 5-8: Result of a Presentation Style 122
Figure 5-9: Result of Phrasal Presenter 131
Figure 5-10: Before Curve Recognition 133
Figure 5-11: After Curve Recognition 134
Figure 6-1: Icon-Style Interface 144
Figure 6-2: Icon-Style Interface 145
Figure 6-3: Icon-Style Interface 147
Figure 6-4: Icon-Style Interface 148
Figure 6-5: Icon-Style Interface 149
Figure 6-6: Icon-Style Interface 151
Figure 6-7: Icon-Style Interface 152
Figure 6-8: Icon-Style Interface 153
Figure 6-9: Menu-Style Interface 155
Figure 6-10: Menu-Style Interface 156
Figure 6-11: Menu-Style Interface 158
Figure 6-12: Menu-Style Interface 159
Figure 6-13: Menu-Style Interface 160
Figure 6-14: Menu-Style Interface 161
Figure 6-15: Menu-Style Interface 163
Figure 6-16: Annotation-Style Interface 165
Figure 6-17: Annotation-Style Interface 166
Figure 6-18: Annotation-Style Interface 168
Figure 6-19: Application Data Base Management 171
