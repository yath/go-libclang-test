#include <clang-c/Index.h>
#include <stdlib.h>
#include "_cgo_export.h"

void myFree(void *p)  {
    free(p);
}

void parseAndVisit(const char *fn, void *cbKey) {
    CXIndex index = clang_createIndex(1, 1);
    CXTranslationUnit unit = clang_parseTranslationUnit(index, fn,
            NULL /* args */, 0 /* n args */,
            NULL /* deleted files */, 0 /* n deleted files */,
            0 /* flags */);
    CXCursor cursor = clang_getTranslationUnitCursor(unit);
    
    clang_visitChildren(cursor, goVisitor, cbKey);
    
    clang_disposeTranslationUnit(unit);
    clang_disposeIndex(index);
}

