import pandas as pd

#BigSudokuDataSetPath='./BigSudokuDataSet.csv'
#ProjectRootPath='D:/CloudComputingLabs/Lab1/src/Sudoku/'
ProjectRootPath='./'
BigSudokuDataSetPath=ProjectRootPath+'BigSudokuDataSet100M.csv'
BigSudokuDataSetInputDF=pd.read_csv(BigSudokuDataSetPath)

'''BigSudokuDataSetInputDF=BigSudokuDataSetInputDF.sample(585000)
BigSudokuDataSetInputDF.to_csv(ProjectRootPath+'BigSudokuDataSet100M.csv',index=0)'''

def sudokuDataCsvGen():
    easyPuzzle=BigSudokuDataSetInputDF.loc[BigSudokuDataSetInputDF['difficulty']<1,['id','puzzle','solution','clues','difficulty'
    ]]
    difficultPuzzle=BigSudokuDataSetInputDF.loc[BigSudokuDataSetInputDF['difficulty']>4,['id','puzzle','solution','clues','difficulty'
    ]]
    randomPuzzle=BigSudokuDataSetInputDF.loc[BigSudokuDataSetInputDF['difficulty']>1,['id','puzzle','solution','clues','difficulty'
    ]]
    while True:
        puzzleDifficulty = input('input the difficulty of sudoku puzzle dataset:easy,difficult or random?')
        puzzleSize = 0
        try:
            puzzleSize = int(input('input the number of sudoku puzzles to generate: '))
        except Exception as e:
            print(e)
        #选择难度评价:easy、difficult、random
        if(puzzleDifficulty=='easy'):
            easyPuzzleTestData=easyPuzzle.sample(puzzleSize)
            easyPuzzleTestData.to_csv(ProjectRootPath+'easyPuzzle{0}.csv'.format(puzzleSize),index=0)
        if(puzzleDifficulty=='difficult'):
            difficultPuzzleTestData = difficultPuzzle.sample(puzzleSize)
            difficultPuzzleTestData.to_csv(ProjectRootPath + 'difficultPuzzle{0}.csv'.format(puzzleSize), index=0)
        if(puzzleDifficulty=='random'):
            randomPuzzleTestData=randomPuzzle.sample(puzzleSize)
            randomPuzzleTestData.to_csv(ProjectRootPath + 'randomPuzzle{0}.csv'.format(puzzleSize), index=0)

def sudokuDataCsvSplitToTxt(inFilePath,outFileRootPath):
    puzzleName='puzzle_'
    solutionName='solution_'
    csvDataSet=pd.read_csv(inFilePath)
    size=csvDataSet.shape[0]
    csvDataSet_puzzle=csvDataSet['puzzle']
    csvDataSet_solution=csvDataSet['solution']

    csvDataSet_puzzle=csvDataSet_puzzle.str.replace('.','0')
    print(csvDataSet_puzzle)
    csvDataSet_puzzle.to_csv(outFileRootPath+'{0}{1}.txt'.format(puzzleName,size),sep=',',index=False,header=False)
    csvDataSet_solution.to_csv(outFileRootPath + '{0}{1}.txt'.format(solutionName, size), sep=',', index=False,header=False)
#sudokuDataCsvGen()

sudokuDataCsvSplitToTxt(ProjectRootPath+'randomPuzzle100000.csv',outFileRootPath=ProjectRootPath)



