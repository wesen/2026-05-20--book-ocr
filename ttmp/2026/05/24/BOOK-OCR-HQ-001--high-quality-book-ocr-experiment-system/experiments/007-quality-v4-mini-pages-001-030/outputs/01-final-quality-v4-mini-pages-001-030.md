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

This is a revised version of a thesis submitted to the Department of Electrical Engineering
and Computer Science on August 27, 1984, in partial fulfillment of the requirements for the
degree of Doctor of Philosophy.

This report describes research done at the Artificial Intelligence Laboratory of the
Massachusetts Institute of Technology. Support for the laboratory's artificial intelligence
research is provided in part by the Office of Naval Research under Office of Naval
Research contract N00014-75-C-0522, in part by the System Development Foundation, and
in part by Wang Laboratories.


<!-- page:004 -->

# PRESENTATION BASED USER INTERFACES

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

# Acknowledgments

My thesis committee, Carl Hewitt, Dick Waters, and Hal Abelson, have been helpful and encouraging. They have all aided significantly in shaping this thesis and improving its quality.

Norton Greenfeld and Martin Yonke introduced me to the world of the presentation concept. It was while working in their group at BBN that I began to think that the concept could serve to explain what was going on in various user interfaces.

Several people have helped with discussions and suggestions at various stages in the development of the ideas, including Lee Blaine, Ron Brachman, Charles Davis, Jeff Gibbons, Earl Killian, Henry Lieberman, Fanya Montalvo, Chuck Rich, Jan Walker, Bill Woods, and Frank Zdybel.

Dan Halbert and Bruce Roberts provided information and the sample screen images for the Xerox Star and Steamer systems, respectively.


<!-- page:006 -->

Table of Contents

Chapter One: Introduction and Overview..........................................................8
1.1 The Primitive Presentation System Model....................................................9
1.2 Constructing Larger Presentation System Models......................................16
1.3 Describing Presentation Systems..........................................................17
1.4 PSBase: A Presentation System Base......................................................18
1.5 Constructing User Interfaces................................................................20
1.6 Related Work......................................................................................21

Chapter Two: The Primitive Presentation System (PPS) Model......................28
2.1 PPSCalc..............................................................................................28
2.2 The Application Data Base................................................................32
2.3 The Presentation Data Base................................................................35
2.4 The Presentation Editor..........................................................................39
2.5 The Presenter.........................................................................................39
2.6 The Recognizer....................................................................................43
2.7 The Representation Shift Model and Direct Manipulation............48

Chapter Three: Constructing Larger Presentation System Models.............54
3.1 Adding a Planned Data Base................................................................54
3.2 Adding a Data Base of Commands........................................................58
3.3 Adding Interfaces to PPS Components.............................................60
3.4 Shared Screen Space and Presentation Structure............................62
3.5 Concluding Remarks...........................................................................66

Chapter Four: Describing Presentation Systems.......................................67
4.1 Emacs Dired.........................................................................................68
4.2 Zmacs...................................................................................................74
4.3 Xerox Star........................................................................................80
4.4 Steamer.............................................................................................90
4.5 Summary of Structural Features................................................97

Chapter Five: PSBase: A Presentation System Base................................100
5.1 Data Base Mechanisms........................................................................103
5.2 Graphics Redisplay...........................................................................114
5.3 Presentation Editor Functions........................................................115
5.4 Presenter Support...............................................................................115
5.5 Recognizer Support..............................................................................124


<!-- page:007 -->

5.6 Basic Style Packages........................................127
5.7 Summary.....................................................141

Chapter Six: Constructing Presentation Systems...................142

6.1 The User's View of the Three Interfaces.....................142
6.2 Common Implementation Details..............................167
6.3 Icon-Style Interface Implementation.........................173
6.4 Menu-Style Interface Implementation........................178
6.5 Annotation-Style Interface Implementation...................181
6.6 Other Style Possibilities...................................183
6.7 Summary.....................................................184

Chapter Seven: Areas for Further Research.......................187

7.1 PSBase Limitations..........................................187


<!-- page:008 -->

Table of Figures

Figure 1-1: A Rudimentary User Interface ................................... 11
Figure 1-2: The Representation Shift Model ................................ 13
Figure 1-3: The Primitive Presentation System (PPS) Model .................. 15
Figure 1-4: Structure of PSBase ........................................... 19
Figure 2-1: The Primitive Presentation System (PPS) Model .................. 29
Figure 2-2: PPSCalc -- Formula Display ................................... 30
Figure 2-3: PPSCalc -- Value Display ..................................... 30
Figure 2-4: PPSCalc -- After Editing ..................................... 31
Figure 2-5: PPSCalc -- After Recalculation ............................... 31
Figure 2-6: PPSCalc -- New Formulas ...................................... 31
Figure 2-7: PPSCalc -- Values of New Formulas ............................ 32
Figure 2-8: World Model ................................................... 34
Figure 2-9: Presenter Parts .............................................. 40
Figure 2-10: Recognizer Parts ............................................ 44
Figure 2-11: PPSCalc -- Value Moved ..................................... 45
Figure 2-12: PPSCalc -- Formula Moved ................................... 46
Figure 2-13: PPSCalc -- Preparing to Copy Formula ........................ 46
Figure 2-14: Representation Shift Model .................................. 49
Figure 2-15: Functional Mapping in the PPS Model .......................... 52
Figure 3-1: Planned Data Base Extension .................................. 56
Figure 3-2: Extension with Both Planning and Immediate Changes ........... 57
Figure 3-3: Command Data Base Extension .................................. 59
Figure 3-4: Presenter Interface Extension ................................ 61
Figure 3-5: Presenter Commands Extension ................................ 63
Figure 4-1: Dired Model .................................................. 72
Figure 4-2: Zmacs Model .................................................. 75
Figure 4-3: Zmacs Scroll Bar ............................................ 81
Figure 4-4: Xerox Star -- Desktop Display ................................. 83
Figure 4-5: Xerox Star -- Opened Folder .................................. 84
Figure 4-6: Xerox Star -- Property Sheet ................................ 86
Figure 4-7: Xerox Star -- Delete Confirmation ............................ 87
Figure 4-8: Xerox Star Model ............................................. 88
Figure 4-9: Sample Steamer Schematic .................................... 91
Figure 4-10: Steamer Menu Console ....................................... 93
Figure 4-11: Steamer Model ............................................... 94
Figure 4-12: Sample of Steamer Icons .................................... 95
Figure 5-1: PSBase Support of PPS Components ............................ 101
Figure 5-2: Structure of PSBase ......................................... 102
Figure 5-3: A Class Description Network ................................ 105


<!-- page:009 -->

Figure 5-4: Sample Presentation Data Base Structure........................................107
Figure 5-5: Inter-Presentation Relationships................................................108
Figure 5-6: Command Description Support......................................................110
Figure 5-7: Reference Resolution................................................................113
Figure 5-8: Result of a Presentation Style....................................................122
Figure 5-9: Result of Phrasal Presenter........................................................131
Figure 5-10: Before Curve Recognition........................................................133
Figure 5-11: After Curve Recognition..........................................................134
Figure 6-1: Icon-Style Interface................................................................144
Figure 6-2: Icon-Style Interface................................................................145
Figure 6-3: Icon-Style Interface................................................................147
Figure 6-4: Icon-Style Interface................................................................148
Figure 6-5: Icon-Style Interface................................................................149
Figure 6-6: Icon-Style Interface................................................................151
Figure 6-7: Icon-Style Interface................................................................152
Figure 6-8: Icon-Style Interface................................................................153
Figure 6-9: Menu-Style Interface................................................................155
Figure 6-10: Menu-Style Interface...............................................................156
Figure 6-11: Menu-Style Interface...............................................................158
Figure 6-12: Menu-Style Interface...............................................................159
Figure 6-13: Menu-Style Interface...............................................................160
Figure 6-14: Menu-Style Interface...............................................................161
Figure 6-15: Menu-Style Interface...............................................................163
Figure 6-16: Annotation-Style Interface........................................................165
Figure 6-17: Annotation-Style Interface........................................................166
Figure 6-18: Annotation-Style Interface........................................................168
Figure 6-19: Application Data Base Management.............................................171


<!-- page:010 -->

Chapter One

Introduction and Overview

Building good user interfaces is a slow and difficult process. Good user interfaces are generally large, complex, and hard to understand, and these characteristics tend to be exacerbated when the interface is modified. All too often, interfaces are built that lack flexibility in their use, lack some functionality, or lack uniformity with interfaces to different applications.

The primary result of this research is the development of a prototype presentation system base, called PSBase. PSBase contains tools, mechanisms, and ready-made parts for the construction of user interfaces. Independence of particular interface styles and application domains is emphasized, in order to maximize the generality and utility of the base. PSBase also provides a conceptual framework for user interfaces. Underlying the base is a general model of user interfaces, called the presentation system model. The report claims that, with a presentation base, interface construction is easier and quicker, and the results are better.

To demonstrate the utility of PSBase, a user interface was constructed on top of it, and three different styles were implemented for this interface. A presentation system base should be independent of any particular application domain or any particular interface style. It should support the construction of (and experimentation with) many different kinds of applications and styles.

For example, consider the following spectrum of styles. At one end is direct manipulation [Shneiderman 83]: the object of interest is continually displayed, and the user's actions appear to be manipulating the object with no intervening command language. An alternative style is preparing a desired future version. (This style looks the same as direct manipulation, but the object of interest is not continually changing -- the specification of the future version is.). Another style is annotating the current view with commands for how to


<!-- page:011 -->

change the object. At the other extreme from direct manipulation is a separate command
language for describing the manipulation. Examples of these alternative styles can be seen
when readers request changes in a draft paper: sometimes the original file is edited,
sometimes a new file is created, sometimes the (paper) draft is annotated, and sometimes the
changes are discussed separately.

Another result of this research is the presentation system model itself. This is a general
model of user interfaces, and it is the foundation of PSBase. Even by itself, however, it has
benefits. It aids the understanding of user interfaces in general by providing a unifying set
of concepts for thinking about user interfaces. There are two ways that it helps someone
building a user interface in the absence of a presentation system base. It serves as a checklist
of the possible kinds of functionality in a user interface. The structure of the model serves
as an architectural framework for the interface.

The model may also be of aid to people studying interface styles in general. One problem
in such a study is the large number and diversity of possible styles. The model defines
various classes of general parameters for interfaces. One can define styles as patterns of
these parameter specifications.

The following five sections provide an overview of the five major chapters in this report.
These chapters divide into two groups. The first group, comprising chapters two, three, and
four, discusses the presentation system model that underlies the presentation system base.
The second group, comprising chapters five and six, discusses the presentation system base
and its application.

1.1 The Primitive Presentation System Model

This section introduces the primitive presentation system (PPS) model of user interfaces,
which is discussed further in chapter two. Two simple models of a data base interface will
first be introduced. They will be used to focus attention on certain aspects and to motivate
the development of the full PPS model. The first model focuses on the data base,
considering a rudimentary interface to it. The second model, the representation shift model,


<!-- page:012 -->

focuses on the user's need for a more useful and coherent representation of the data base information and commands. The representation shift model is also useful in itself, as it is a special case of the full PPS model and applies to some common interface styles. The PPS model extends the representation shift model to allow more flexibility in the relationship between the screen and the data base.

A Rudimentary User Interface. Figure 1-1 shows the basic interface to an application data base and a rudimentary user interface constructed from it. The data base has three external inputs and outputs. Commands change the state of the data base (adding, changing, or deleting information). Queries allow the state of the data base to be examined, producing the relevant information at the observables output.

These inputs and outputs are not directly usable by a person -- they are in a format designed for use by programs. (The user is not the only one using data bases, after all.) In order to provide even a rudimentary user interface, some simple kind of transducers must be placed on each input and output line.

The transducer on the command input, for example, might convert a text version of a command to the binary form required by the data base. The transducers do not provide a different overall model of data base use -- the user still must use the commands and queries provided by the data base. The language used to express them has been changed slightly so that it is printable and mnemonic, much the same kind of translation that a simple assembler performs.

The rudimentary interface is usable, but suffers from two basic problems from the user's point of view. First, the user must express the data base modification in terms of the data base commands available. Second, the results of such modification, as well as any viewing desired, must be explicitly requested via queries.

Representation Shift. Figure 1-2 shows an expanded user interface. Here, two data bases are involved, the application data base as before and a new one, called a presentation of the data base, introduced to allow the user more direct modification and viewing. The


<!-- page:013 -->

Figure 1-1: A Rudimentary User Interface

[FIGURE: Diagram showing users represented by circles labeled "T" connected to an Application Data Base with arrows labeled "queries", "observables", and "commands"]


<!-- page:014 -->

presentation data base contains the same information as the application data base, but it is
represented in a way that is directly viewable, i.e., in terms of text and graphic forms. It is
continuously displayed (on the user's terminal), so that the user does not have to explicitly
request information to be viewed.

The presentation -- or, loosely speaking, the screen -- can be directly edited by the user,
by means of the presentation editor. The editor allows the user to manipulate the forms on
the screen, creating new forms or changing or deleting existing ones. Conceptually, it
combines capabilities of a text editor with those of a graphics (diagram) editor. As these
changes are made, their results are immediately visible.

In addition, the commands for presentation editing are chosen to be convenient for the
user. For example, they might include a base of general text-editing and graphics-editing
commands, so that the user does not have to learn a special language for each particular
application data base.

The presenter creates the presentation data base from the application data base. At
appropriate times as the user edits the presentations, the recognizer creates a new version of
the application data base from the presentation data base. In the representation shift model
the presentation contains all and only the information contained in the application data
base. The presenter uses a single application data base query (labeled get-db in the figure)
to get a representation of the entire application data base, converts the representation, and

the presentation data base. Similarly, the recognizer gets the entire presentation contents,
converts it, and loads the entire application data base.

In the representation shift model, the presenter relation must be invertible, since the
recognizer must be able to specify the entire application data base from the presentation
data base. In general the presenter relation is a subset of the recognizer relation, or in other
words, the recognizer will recognize several different variants of the same presentation,
allowing the user more latitude. For example, the recognizer might allow the user to create
any of "12", "12.0", "12.000", etc., whereas the presenter might always choose "12.0".


<!-- page:015 -->

Figure 1-2: The Representation Shift Model

User -> Presentation Editor
editing commands

Presentation Editor -> Presentation Data Base

Presentation Data Base

Command (LOAD-DB) -> Presenter

Presenter <- query (GET-DB) - - - -> Application Data Base
All info in DB

Presentation Data Base - - - query (GET-DB) -> Recognizer
All info in DB -> Recognizer

Recognizer -> Application Data Base
Command (LOAD-DB)


<!-- page:016 -->

The representation shift model is a direct manipulation interface [Shneiderman 83]. The screen continuously displays the data base. Whenever the data base changes, the screen is updated. Similarly, the user manipulates the data base by manipulating the forms on the screen, and the data base is continually updated from this.

The major restriction of the representation shift model is that the entire application data base be viewed (and in an invertible presentation). This can lead to inefficiency. It can also lead to the inconvenience of visual clutter -- the user cannot view just a relevant subset of a complex data base. The ability to control the selection of information to be viewed and the way it is to be viewed can be crucial to the successful use of the data base.

**The Full PPS Model.** The full PPS model, shown in figure 1-3, relaxes the restriction that the entire application data base must be viewed. The presentation, i.e., the visual data base, may convey only a small part of the information in the application data base. The screen thus can no longer be recognized in a simple manner as specifying all the information in the application data base. This necessitates a generalization in the recognizer from that in the representation shift model: the recognizer translates editing actions into data base commands, rather than translating editing results into data base data. (The term editing actions includes both the editing command and the editing result. Therefore, the PPS recognizer includes, as a special case, the possibility of just having to examine the editing result.)

The presenter is responsible for making the screen continually show the relevant part of the data base. It creates the initial display and updates the display when the data base changes. The presenter collects the relevant information from the application data base, converts that information to text and/or graphics, and organizes the layout of this visual information on the screen.

The recognizer causes the data base to change to reflect the user's editing of the presentation. Specifically, in addition to affecting the screen, the user's editing operations are recognized as -- i.e., translated into -- operations on the data base. Thus, the PPS model is also a direct manipulation interface: the data base is continually presented on the screen,


<!-- page:017 -->

Figure 1-3: The Primitive Presentation System (PPS) Model

User -> Presentation Editor
editing commands

Presentation Editor -> Presentation Data Base
Presentation Data Base

Presenter
data base commands

Presenter -> Application Data Base
observables

Application Data Base
data base commands

Presentation Data Base -> Registrar
editing actions

Registrar


<!-- page:018 -->

with screen following data base changes (by presenter action) and data base following screen
changes (by recognizer action).

1.2 Constructing Larger Presentation System Models

The primitive presentation system model can be extended to model more complex
presentation systems as discussed in chapter three. The basic technique for extending the
presentation system model is to attach an additional presentation system to it, either
replacing or augmenting some part of it. The resulting presentation system may thus
contain several smaller presentation systems. The extensions discussed in this section are
suggested by examining the major limitations of the PPS model.

Adding a Planned Data Base. In the PPS model changes to the data base are immediate.
To avoid this, a second application data base can be added to a presentation system: a
future (i.e., planned) version of the original data base. The user can edit the planned
version's presentation, separate from the presentation of the current state of the data base.
When the planned version looks acceptable, the user gives a do it command that causes the
actual data base to be updated.

Adding a Data Base of Commands. In the PPS model the user cannot see a description of
the changes or the commands to effect them presented explicitly. (Only the data base that
results from these commands is seen.) Using a technique similar to the previous one of
adding a planned version of the data base, a data base of commands can be added. In this
extension, the planned changes are represented in the new data base explicitly, and can be
presented in a style different from the style for the application data base.

Adding Interfaces to PPS Components. In the PPS model the editor, presenter, and
recognizer are not presented; the user only has an interface of primitive signals to them
(e.g., keystrokes or a pointing device). To circumvent this limitation, presentation system
interfaces to these components can be added. One technique involves adding a data base
for the particular component's state, e.g., some options controlling the presenter's style, and
constructing presenters and recognizers for showing and manipulating it. Alternatively, a


<!-- page:019 -->

data base of commands for the component can be added, just as in the previous section a
command data base was added for the application data base.

1.3 Describing Presentation Systems

The presentation system model can be used as a descriptive tool.  The model provides a
set of concepts for enumerating and categorizing basic functions and interactions in a user
interface, even when that interface was not designed with the model in mind.

In chapter four several user interfaces will be described using the presentation system
model.  The selection exhibits a variety of interface styles in order to illustrate the model's
generality.  In each example the focus will be on those presentation system mechanisms that
play the most important part in defining that particular style.  Two interfaces, drawn from
those described in chapter four, are sketched below.

Xerox Star / Apple Lisa.  The Xerox Star [Smith, Irby, Kimball, Verplank & Harslem 83]
and the Apple Lisa [Lisa 84] systems offer an interface using icons -- pictorial presentations
of commands and data.  Some recognition is simple reference resolution such as pointing to
an icon that presents a particular command.  Other recognition involves more complicated
inter-icon relations such as proximity.  For example, in Lisa the user deletes a file by
moving the file's icon to a trash can icon.  In both Star and Lisa the user prints the file by
moving its icon to the printer icon.

Emacs Dired.  A subsystem of the Emacs editor [Stallman 81], Dired is used to perform
various directory operations.  It is an example of an extended presentation system that
provides both direct manipulation of the data base (the directory being edited), e.g., when
certain file properties are changed, and planned operations, e.g., when files are marked for
later deletion.  The planned deletions are presented as annotations to the presentation of the
current directory.


<!-- page:020 -->

# 1.4 PSBase: A Presentation System Base

Chapter five discusses PSBase, the prototype presentation system base that was implemented in the course of this research.

PSBase explicitly incorporates the presentation system model structure. It includes tools, mechanisms, and ready-made parts for building an interface consisting of an application data base, presentation data base, presenters, recognizers, and presentation editor. Domain-independent and style-independent mechanisms are provided and may be combined largely independently. These characteristics cause PSBase to be useful in constructing a wide range of interfaces.

Figure 1-4 shows the overall structure of PSBase. The data base mechanisms provide support for building application data bases structured in a network somewhat similar to knowledge representation networks. The network includes descriptions of the classes of objects as well as the objects themselves, and class inheritance is supported. An important point is that this network is used to build the presentation data base as well, and the presentation and application data bases are linked together into a large, uniformly structured data base. This uniformity is an important factor in the power of the PSBase mechanisms. PSBase predefines a large part of the presentation data base class network.

PSBase also provides mechanisms that accompany the presentation data base: Graphics redisplay ensures that the presentation data base is continuously displayed on the terminal. Several presentation editor functions are provided; the interface builder may select these, as desired.

The presenter support and recognizer support modules provide a variety of tools and mechanisms for creating and controlling presenters and recognizers. Most important among these mechanisms is a language for describing presentation styles and general presenters that interpret these languages. The interface builder need only describe how the presentation structure relates to the application data base structure, and the presenters perform the actual creation and updating of the presentations.


<!-- page:021 -->

Figure 1-4: Structure of PSBase

BASIC STYLE PACKAGES

PRESENTER
SUPPORT

RECOGNIZER
SUPPORT

GRAPHICS
REDISPLAY

EDITOR
FUNCTIONS

DATA BASE MECHANISMS

[FIGURE: Structure of PSBase diagram]


<!-- page:022 -->

A number of basic style packages offer specific components of domain-independent
interface styles that the interface builder may choose to include. Some general presenters
and recognizers are provided. For example, a presenter is provided to produce command
menus. As another example, a recognizer is provided to interpret simple rule descriptions
in order to recognize icon movement, similar to the Xerox Star and Apple Lisa systems (see
section 1.3).

No claim is made that PSBase would serve as a production presentation system base. It is
a prototype, and needs more and improved features of many kinds. It provides only a part
of the presentation editor functions that would be needed. Many more domain-
independent presenters and recognizers could be included. The presentation style
description language could be improved and used to drive recognition as well. This would
result in more uniformity in what the system can present and what it can recognize,
providing the user with increased consistency and power.

1.5 Constructing User Interfaces

In order to demonstrate the utility of PSBase, three interfaces were constructed using the
PSBase mechanisms and tools. The three interfaces share the same application data base,
but embody different styles. The first style uses icons, similar to the Xerox Star and Apple
Lisa system described in section 1.3. The second style uses text displays with accompanying
command menus. The third style is a graphical annotation style, an extension of the Dired
style described in section 1.3.

Some of the work was done once and shared between the three implementations, namely,
the style-independent development of the application data base. Once that work was
completed, implementing a particular style was largely a matter of writing a few small pieces
using PSBase tools and choosing some standard PSBase ready-made parts from the basic
style packages module.


<!-- page:023 -->

1.6 Related Work

This report discusses two developments, a domain-independent, style-independent presentation system base for building user interfaces, and its underlying model of user interfaces. This section discusses characteristics of the base and the model that distinguish it from other research. Two characteristics of both the base and the model are particularly important:

First, the model and the base attempt to concentrate on general mechanisms, independent of any particular domain and independent of any particular style. The intent has been that they should be free of value judgments concerning styles. Discussing what constitutes a good style or developing new styles are separate efforts; this research offers a conceptual vocabulary in which such a discussion can be phrased and offers a base for experimenting with or combining alternative styles.

Second, the model and the base center about the high-level concept of the presentation.

This concept considers the semantic connection between the screen and the application.
The model is structured to show how the presentation is used as a medium for communication between the user and the application. The emphasis in both the model and the presentation system base has been on the system aspects: how the system of processes and data bases are structured and interact regarding the presentation relationship. This research has not emphasized any one particular part of this system: several other studies emphasize the application data base, or the presentation data base, or presenters, or recognizers.

Other research that this work resembles can be classed into three broad areas: human factors, systems and techniques, and presentation systems. Although this research is related to these areas, the author knows of no other research that directly addresses the same goals of studying and providing support for a system of general user interfaces mechanisms. Rather than being an alternative approach, this work complements the others that are mentioned. The third area, presentation systems, is the closest to this research, in that its includes systems for aiding user interface construction, based on concepts similar to the


<!-- page:024 -->

presentation concept used here.

Human Factors. At the psychological end of the spectrum, there have been several efforts to which this research is somewhat related. Two major kinds of work is described, first, user modeling and, second, interface specification techniques and guidelines. Some representative research is mentioned.

There have been efforts to develop models of user behavior, user performance, and user understanding of systems. Often these studies concentrate on particular classes of users or interface styles. Shneiderman, for example, has examined a class of interface styles that he terms direct manipulation [Shneiderman 83]. These interfaces are marked by "visibility of the object of interest; rapid, reversible, incremental actions; and replacement of complex command language syntax by direct manipulation of the object of interest." He discusses direct manipulation style, and its affect on and acceptance by different kinds of users, in terms of a semantic/syntactic model of user behavior [Shneiderman & Mayer 79] [Shneiderman 80]. According to this model, two kinds of knowledge about user interfaces reside in long-term memory, syntactic and semantic. Syntactic knowledge includes details of command syntax; it has an arbitrary character and is easily forgotten unless frequently used. Semantic knowledge includes the hierarchically-structured concepts of functionality and processes for performing various tasks. Semantic knowledge is largely independent of particular systems and is more easily retained. The success of the direct manipulation style follows from the fact that "the object of interest is displayed so that actions are directly in the high-level problem domain," requiring little need for syntactic knowledge.

Modeling the user can be a tool for evaluating the behavioral style of an interface, by studying the match between the interface behavior and the user behavior. The presentation system model, on the other hand, complements the user model by approaching the problem from the other end, discussing the kinds and internal structures of interface mechanisms that will by their interaction produce the particular overall behavior as seen by the user.

Some guidelines and formal techniques have been developed for specifying user interface dialogs, a part of the user interface style. Formal grammars (or, equivalently, state transition


<!-- page:025 -->

networks) are one technique for describing and designing the dialog between user and computer [Reisner 81][Reisner 82][Bleser & Foley 82][Jacob 82][Brown 82]. Formal grammars describe the interaction between user actions and system responses. Some grammars include cognitive information, describing what a user has to learn and remember. A grammar can be used as a design tool, evaluating designs for consistency and simplicity. Problems users might have and mistakes they might make can be predicted.

As with user models, dialog descriptions are complemented by the work reported here. One may identify three layers of study, all requiring models and description techniques: general user interface mechanisms (presentation system model), overall user interface style (dialog specifications), and the user (user models).

Systems and Techniques. The second area of related work is the building of systems, from cooperative user interfaces to graphics systems, and the development of techniques to use in such systems. Some of these projects tend to concentrate on one side or the other of the presentation relation: on representing the knowledge in the application data base or on manipulating and displaying the presentation data base. Others tend to concentrate on the development of particular interface styles.

Research into cooperative user interfaces, such as the Cousin effort at CMU [Hayes 84] and the Consul/Cue effort at Information Sciences Institute [Kaczmarek, Mark & Wilczynski 83][Mark 81], study various ways that user interface can be more easily constructed to actively aid the user. An important part of such systems is the provision of a uniform view of the applications and a helpful assistant, based on an extensive description of those applications or the interface styles. Such an assistant might try to understand why the user is having difficulty or try to understand requests made in an unexpected form.

A large part of the Consul/Cue work concentrates on the representation of knowledge about the application and its commands (services). The different applications are described in a uniform manner. This is separated from the particular choice of styles used to interface to these applications, such as windows/pointing, command languages, or natural language. The user interface assistant understands the data base representation and uses it to provide


<!-- page:026 -->

explanations, flexible recovery from command language errors, and assistance in using several different applications by understanding their functionality.

The research reported here is closer to the Cousin project. The Cousin project does not concentrate on incorporating knowledge about application semantics, but rather on developing a uniform interface style to support a user interface assistant. The assistant corrects erroneous or abbreviated input, interacts with the user to resolve errors, and offers integral and automatically generated on-line help and documentation. The Cousin system provides a common interface base, separate from the application, that interprets an interface definition provided by the application builder. This definition expresses the user interface as a set of forms, with fields that convey information between the user and the application.

There is an emphasis in these research efforts on developing cooperative styles, developing techniques for them (such as more intelligent recognizers), and for Consul/Cue, investigating the problems of representing knowledge about the application's functionality. The work reported in this report also relies heavily on the separation and uniformity of the application data base mechanism. But this work has not studied the issues of knowledge representation involved. Nor has it been involved with developing particular styles. And unlike the cooperative systems projects, this work attempts to be able to model and support arbitrary existing interface styles.

There are several research efforts studying different uniform styles of information presentation and interaction, and several efforts at developing presentation and interaction techniques for specific domains. For example, spatial data base management systems [Herot 80] [Donelson 78], the Boxer system [diSessa 85], the Xerox Star [Purvy, Farrell & Klose 83] [Smith, Irby, Kimball, Verplank & Harslem 83], and the Query-by-Example-based office systems [Zloof 82] [Zloof & de Jong 77] all offer the user a consistent way of interacting with a variety of applications. In a spatial data base management system, the user accesses information by "moving through" the data base -- information from many different domains is organized spatially, with related information nearby. Retrieval is something like flying over a land of information: information is found by moving to it, and


<!-- page:027 -->

detail is controlled by zooming. In the Query-by-Example systems, on the other hand, the
user accesses different kinds of information by providing an example of the kind of
information desired. Several systems have been developed that offer complex presentation
techniques and styles for particular domains. Simulators are perhaps the most widely
known; the Steamer system [Stevens, Roberts & Stead 83], discussed in chapter four, is one
example. Another area of increasing interest is the presentation of the organization and
execution of programs, such as the Computer Corporation of America's program
visualization system [CCA 79], Henry Lieberman's Tinker system [Lieberman 84]
[Lieberman 83], and the Brown University system for program animation [Brown &
Sedgewick 84a][Brown & Sedgewick 84b][Brown & Sedgewick 84c]. The intent of the
work reported in this report is to develop a model and system that can be used to describe
and build any of these kinds of styles.

The books by Newman and Sproull [Newman & Sproull 79] and Foley and Van Dam
[Foley & Van Dam 82] primarily discuss low-level drawing and interaction techniques for
graphics systems. For the most part, they are concerned with only one kind of application
data base -- geometric models of solids, surfaces, etc. Within the framework of the model of
this report, their books discuss detailed techniques for building presentation editors and
presentation data bases. However, concerning the presentation data base, their emphasis is
more on representation at a low level, suitable for display processors, and does not attempt
to offer a general representation technique. This is in contrast to the presentation system
base of chapter five, for example, which uses a general description mechanism for both the
presentation data base and the application data base. The standard graphics systems are less
in need of such a scheme, as they are not involved with any sort of "reasoning" about the
data bases, and instead need to perform computations efficiently. Thus, the graphics system
should be viewed as a low-level component of a presentation data base as described in this
report.

Information Presentation Systems. The research reported in this report most closely
resembles research developing what have been called information presentation systems or
systems for automatically synthesizing graphics environments, for example the Bharat system


<!-- page:028 -->

[Gnanamgari 81], the View system [Friedell 83], and the AIPS system [Zdybel, Gibbons,
Greenfeld & Yonke 81][Zdybel, Greenfeld, Yonke & Gibbons 81]. These systems all
emphasize a knowledge-based approach to creating what this report would call intelligent
presenters. The systems explicitly incorporate concepts similar to the presentation concept
used here, particularly the AIPS system. All three systems have interesting and individual
aspects, but from the point of view of this research, it will suffice to discuss the AIPS work
as representative. (It was while working with the AIPS group that the author first started
thinking about the presentation's use as an organizing concept for modeling user interfaces.)

The goal of AIPS as an information presentation system is to provide an interface to a
large knowledge base or knowledge-based system. The system automatically generates
displays from content-oriented (i.e., domain) specifications. (E.g., "display the ships in the
Mediterranean.") AIPS is itself a knowledge-based system. Using a large knowledge base
describing how structures of domain information can be related to structures of graphical
displays, the system automatically selects or constructs an appropriate presentation style. A
full information presentation system would include knowledge about the user, general
domains, a wide variety of presentation styles, and human factors decisions involved in
graphical display.

There are three aspects in which the work reported in this report differs from the AIPS
research. First, this report addresses a more general class of interfaces than information
presentation systems. Information presentation systems currently exist only in prototype
form; there are many other kinds of interfaces to be supported now and, presumably, even
when full information presentation systems are available. Most interfaces do not have
intelligent or automatic presenters. One reflection of this difference is seen in the general
model of interfaces developed in this report.

Second, this report emphasizes the system aspects of the interface, rather than
concentrating on any one component of the system. This is one reason why this research
and the others are complementary: the AIPS work considers presenters in detail; this work
considers the relationship between presenters and the rest of the user interface system.


<!-- page:029 -->

Third, the most distinguishing characteristic of the AIPS work is its emphasis on issues of knowledge representation. This report does not address those issues, again because the emphasis here is not on intelligent presenters or on techniques of describing presentation styles. Relatively simple description techniques suffice for the PSBase system. However, the results of research into the representation of knowledge about graphical display could be incorporated into a production version of a presentation system base to great effect.


<!-- page:030 -->

# Chapter Two

# The Primitive Presentation System (PPS) Model

This chapter discusses the PPS model in detail. Figure 2-1 reproduces figure 1-3 of chapter one, except that here two new primitive-signal inputs are added, controls for the presenter and recognizer. Each of the components of the PPS model will be discussed in turn in sections below.

## 2.1 PPSCalc

The sections in this chapter use an example program called PPSCalc. This is a simple spreadsheet program, a trivial version of VisiCalc [Beil 82]. PPSCalc was designed specifically for this explanation -- its behavior strictly follows the PPS model. PPSCalc is illustrated in figures 2-2 and 2-3.

The spreadsheet consists of cells, organized in rows and columns. Each cell may be empty, contain just a numeric value, or contain a formula and a numeric value. In a cell with a formula, the numeric value is computed by the formula from the values in other cells. Cells which just have a numeric value -- no formula -- are called independent cells. Their values are set by the user. Cells which have a formula are called dependent cells. Their values are recomputed periodically, as will be discussed below. Cells with neither a formula nor a value are empty.

PPSCalc has two display modes, formula display and value display, illustrated by the two figures. Figure 2-2 shows the mode that displays the dependent cells' formulas. Figure 2-3 shows the mode displaying the dependent cells' values computed by those formulas.

PPSCalc is shown in figure 2-2 with an assignment of cell formulas for computing a simple bill, based on the prices for two kinds of items and the numbers of the items
